// A CLI guardian, not a UI automation tool. The original clipboard stays local
// in a bounded, protected binary plist; it is never printed or sent to a review.
import AppKit
import Foundation
import UniformTypeIdentifiers
import Darwin

enum Refusal: Error { case bounds, changed, unreadable, privateFile, journal, ownership, childLive, mismatch, arguments }
struct Flavor: Codable, Equatable { let type: String; let data: Data }
struct Snapshot: Codable, Equatable { let aliases: [Flavor]; let items: [[Flavor]] }
struct State: Codable {
    let version: Int; let nonce: String; let phase: String; let kind: String
    let revision: Int; let itemCount: Int
}
let markerMIME = "application/x-vellum-clipboard-test-owner"
let legacyPrefix = "Vellum legacy clipboard check "
let legacyStringType = NSPasteboard.PasteboardType("NSStringPboardType")
let maxBackup = 20 << 20
let fm = FileManager.default

func revision(_ board: NSPasteboard) throws -> Int {
    let value = VellumGuardRevision(board)
    if value < 0 { throw Refusal.unreadable }; return value
}
func boardItems(_ board: NSPasteboard) throws -> [NSPasteboardItem] {
    guard let items = VellumGuardItems(board) else { throw Refusal.unreadable }; return items
}
func boardTypes(_ board: NSPasteboard) throws -> [NSPasteboard.PasteboardType] {
    guard let types = VellumGuardBoardTypes(board) else { throw Refusal.unreadable }; return types
}
func itemTypes(_ item: NSPasteboardItem) throws -> [NSPasteboard.PasteboardType] {
    guard let types = VellumGuardItemTypes(item) else { throw Refusal.unreadable }; return types
}
func newItem() throws -> NSPasteboardItem {
    guard let item = VellumGuardItem() else { throw Refusal.unreadable }; return item
}
@discardableResult func clear(_ board: NSPasteboard) throws -> Int {
    let value=VellumGuardClear(board)
    if value < 0 { throw Refusal.unreadable }; return value
}

// Snapshot/materialization is complete before any test mutation. Limits cannot
// bound time/memory inside a lazy OS provider; refusal never starts the fixture.
func capture(_ board: NSPasteboard, declaredTypes: (() throws -> [NSPasteboard.PasteboardType])? = nil) throws -> Snapshot {
    func declared() throws -> [NSPasteboard.PasteboardType] {
        if let declaredTypes { return try declaredTypes() }; return try boardTypes(board)
    }
    let initial = try revision(board)
    let native = try boardItems(board)
    let boardDeclared=try declared()
    if native.count > 32 || boardDeclared.count>1024 || Set(boardDeclared).count != boardDeclared.count { throw Refusal.bounds }
    var items: [[Flavor]] = []; var total = 0; var metadata = 0
    var represented=Set<NSPasteboard.PasteboardType>()
    for item in native {
        let types = try itemTypes(item)
        if types.isEmpty || types.count > 32 || Set(types).count != types.count { throw Refusal.bounds }
        var flavors: [Flavor] = []
        for type in types {
            let length = type.rawValue.utf8.count
            if length == 0 || length > 1024 || type.rawValue.utf8.contains(0) || length > 65536 - metadata { throw Refusal.bounds }
            metadata += length
            guard let data = VellumGuardData(item, type) else { throw Refusal.unreadable }
            if data.count > 8 << 20 || data.count > (16 << 20) - total { throw Refusal.bounds }
            total += data.count
            flavors.append(Flavor(type: type.rawValue, data: Data(data)))
        }
        if try itemTypes(item) != types { throw Refusal.changed }
        represented.formUnion(types)
        items.append(flavors)
    }
    let declaredSet=Set(boardDeclared), missing=declaredSet.subtracting(represented)
    if !represented.isSubset(of:declaredSet) {throw Refusal.unreadable}
    var aliases:[Flavor]=[]
    if !missing.isEmpty {
        // Cocoa advertises this derived legacy alias outside item.types. Accept
        // only exact first-item canonical bytes; capture those raw alias bytes
        // too and require their exact reconstruction after original restoration.
        guard missing==Set([legacyStringType]), let canonical=items.first?.first(where:{$0.type==NSPasteboard.PasteboardType.string.rawValue}), let raw=VellumGuardBoardData(board,legacyStringType), raw==canonical.data else {throw Refusal.unreadable}
        let length=legacyStringType.rawValue.utf8.count
        if length>65536-metadata || raw.count>8<<20 || raw.count>(16<<20)-total {throw Refusal.bounds}
        metadata+=length;total+=raw.count
        aliases.append(Flavor(type:legacyStringType.rawValue,data:Data(raw)))
    }
    if try revision(board) != initial || declared() != boardDeclared { throw Refusal.changed }
    return Snapshot(aliases:aliases,items: items)
}
func nativeItems(_ snapshot: Snapshot) throws -> [NSPasteboardItem] {
    try snapshot.items.map { flavors in
        let item = try newItem()
        for flavor in flavors {
            if !VellumGuardSetData(item, flavor.data, .init(flavor.type)) { throw Refusal.unreadable }
        }
        if try itemTypes(item).map(\.rawValue) != flavors.map(\.type) { throw Refusal.mismatch }
        return item
    }
}
func privateDirectory(_ path: String) throws {
    var info = stat()
    if lstat(path, &info) != 0 || (info.st_mode & S_IFMT) != S_IFDIR || info.st_mode & 0o777 != 0o700 || info.st_uid != geteuid() { throw Refusal.privateFile }
}
func privateData(_ path: String, maximum: Int) throws -> Data {
    let fd = open(path, O_RDONLY | O_NOFOLLOW)
    if fd < 0 { throw Refusal.privateFile }; defer { close(fd) }
    var info = stat()
    if fstat(fd, &info) != 0 || info.st_mode & S_IFMT != S_IFREG || info.st_mode & 0o777 != 0o600 || info.st_uid != geteuid() || info.st_size < 0 || info.st_size > maximum { throw Refusal.privateFile }
    var data = Data(count: Int(info.st_size)); var offset = 0
    try data.withUnsafeMutableBytes { raw in
        while offset < raw.count {
            let n = read(fd, raw.baseAddress!.advanced(by: offset), raw.count - offset)
            if n < 0 && errno == EINTR { continue }
            if n <= 0 { throw Refusal.privateFile }; offset += n
        }
    }
    return data
}
// Same-file writers are serialized: parent before launch, child native main
// thread during its lifetime, parent/recovery only after confirmed child exit.
func durable(_ data: Data, _ path: String) throws {
    let directory = (path as NSString).deletingLastPathComponent
    try privateDirectory(directory)
    // A crashed writer's temporary must not block restoration or be deleted.
    var template = Array((path + ".new.XXXXXX").utf8CString)
    let fd = template.withUnsafeMutableBufferPointer { mkstemp($0.baseAddress!) }
    if fd < 0 { throw Refusal.privateFile }
    let temporary = template.withUnsafeBufferPointer { String(cString: $0.baseAddress!) }
    var okay = false
    defer { close(fd); if !okay { unlink(temporary) } }
    try data.withUnsafeBytes { raw in
        var offset = 0
        while offset < raw.count {
            let n = write(fd, raw.baseAddress!.advanced(by: offset), raw.count - offset)
            if n < 0 && errno == EINTR { continue }
            if n <= 0 { throw Refusal.privateFile }; offset += n
        }
    }
    if fsync(fd) != 0 || rename(temporary, path) != 0 { throw Refusal.privateFile }
    let dfd = open(directory, O_RDONLY | O_DIRECTORY | O_NOFOLLOW)
    if dfd < 0 { throw Refusal.privateFile }; defer { close(dfd) }
    if fsync(dfd) != 0 { throw Refusal.privateFile }; okay = true
}
func json<T: Encodable>(_ value: T) throws -> Data { try JSONEncoder().encode(value) }
func state(_ dir: String) throws -> State {
    let record = try JSONDecoder().decode(State.self, from: privateData(dir + "/state.json", maximum: 2048))
    if record.version != 1 || record.revision < 0 || record.itemCount < 0 || record.itemCount > 32 || record.nonce.count != 64 || record.nonce.range(of: "^[0-9a-f]{64}$", options: .regularExpression) == nil { throw Refusal.journal }
    return record
}
func markerType() throws -> NSPasteboard.PasteboardType {
    guard let type = UTType(mimeType: markerMIME) else { throw Refusal.unreadable }
    return .init(type.identifier)
}
func owns(_ board: NSPasteboard, _ record: State) throws -> Bool {
    if try revision(board) != record.revision { return false }
    let items = try boardItems(board)
    if record.phase == "empty" || record.phase == "restore_empty" {
        return try record.itemCount == 0 && items.isEmpty && boardTypes(board).isEmpty && revision(board) == record.revision
    }
    if record.phase != "owned" || items.count != record.itemCount || items.isEmpty { return false }
    var matches = false
    if record.kind == "rich" {
        let type = try markerType()
        matches = try itemTypes(items[0]).contains(type) && VellumGuardData(items[0], type) == Data(record.nonce.utf8)
    } else if record.kind == "legacy" {
        matches = try items.count == 1 && itemTypes(items[0]).contains(.string) && VellumGuardData(items[0], .string) == Data((legacyPrefix + record.nonce).utf8)
    }
    return try matches && revision(board) == record.revision
}
func noLiveChild(_ dir: String) throws {
    let path = dir + "/child.pid"
    if !fm.fileExists(atPath: path) { return } // Fixture cannot write until this durable gate exists.
    let bytes = try privateData(path, maximum: 32)
    guard let text = String(data: bytes, encoding: .ascii), let pid = Int32(text.trimmingCharacters(in: .whitespacesAndNewlines)), pid > 1 else { throw Refusal.journal }
    let deadline = Date().addingTimeInterval(5)
    while true {
        if kill(pid, 0) != 0 && errno == ESRCH { return }
        if Date() >= deadline { throw Refusal.childLive }
        Thread.sleep(forTimeInterval: 0.02)
    }
}
func restore(_ board: NSPasteboard, _ dir: String, _ original: Snapshot, _ record: State, interpose: ((String,NSPasteboard) throws -> Void)? = nil) throws {
    // Recovered "admitted" or "restore_admitted" states do not prove ownership.
    // A process interrupted after restoring may prove exact original bytes without
    // another mutation. Reading those bytes is bounded by the capture policy.
    if try record.phase == "original" && revision(board) == record.revision {
        if try capture(board) != original { throw Refusal.mismatch }; return
    }
    if try record.phase == "restored" && revision(board) == record.revision {
        if try capture(board) != original { throw Refusal.mismatch }; return
    }
    if try !owns(board, record) { throw Refusal.ownership }
    let items = try nativeItems(original) // Entire restored item set prepared before clearing.
    try durable(json(State(version: 1, nonce: record.nonce, phase: "restore_admitted", kind: record.kind, revision: record.revision, itemCount: record.itemCount)), dir + "/state.json")
    try interpose?("admitted",board) // Private-fixture injection only; normal runs use nil.
    // Cocoa provides no atomic compare-and-swap across this check/clear/write.
    if try !owns(board, record) { throw Refusal.ownership }
    let emptyRevision = try clear(board)
    let empty=State(version: 1, nonce: record.nonce, phase: "restore_empty", kind: "empty", revision: emptyRevision, itemCount: 0)
    if try !owns(board,empty) {throw Refusal.ownership}
    try durable(json(empty), dir + "/state.json")
    try interpose?("empty",board)
    if try !owns(board,empty) {throw Refusal.ownership}
    if !items.isEmpty && !VellumGuardWrite(board, items) { throw Refusal.unreadable }
    let finalRevision = try revision(board)
    if try capture(board) != original || revision(board) != finalRevision { throw Refusal.mismatch }
    try durable(json(State(version: 1, nonce: record.nonce, phase: "restored", kind: "", revision: finalRevision, itemCount: original.items.count)), dir + "/state.json")
}
func eraseBackup(_ dir: String) throws {
    if unlink(dir + "/original.plist") != 0 { throw Refusal.privateFile }
    let fd = open(dir, O_RDONLY | O_DIRECTORY | O_NOFOLLOW)
    if fd < 0 { throw Refusal.privateFile }; defer { close(fd) }
    if fsync(fd) != 0 { throw Refusal.privateFile }
}
func loadOriginal(_ dir: String) throws -> Snapshot {
    try privateDirectory(dir)
    let data = try privateData(dir + "/original.plist", maximum: maxBackup)
    let snapshot = try PropertyListDecoder().decode(Snapshot.self, from: data)
    // Validate decoded backup counts, metadata and byte bounds before constructing
    // native items; a corrupt private backup is never interpreted as a write.
    var total = 0; var metadata = 0
    if snapshot.items.count > 32 { throw Refusal.bounds }
    for item in snapshot.items {
        if item.isEmpty || item.count > 32 || Set(item.map(\.type)).count != item.count { throw Refusal.bounds }
        for flavor in item {
            let n = flavor.type.utf8.count
            if n == 0 || n > 1024 || flavor.type.utf8.contains(0) || n > 65536 - metadata || flavor.data.count > 8 << 20 || flavor.data.count > (16 << 20) - total { throw Refusal.bounds }
            metadata += n; total += flavor.data.count
        }
    }
    if snapshot.aliases.count>1 {throw Refusal.bounds}
    for alias in snapshot.aliases {
        guard alias.type==legacyStringType.rawValue, let canonical=snapshot.items.first?.first(where:{$0.type==NSPasteboard.PasteboardType.string.rawValue}), alias.data==canonical.data else {throw Refusal.mismatch}
        let n=alias.type.utf8.count
        if n>65536-metadata || alias.data.count>8<<20 || alias.data.count>(16<<20)-total {throw Refusal.bounds}
        metadata+=n;total+=alias.data.count
    }
    return snapshot
}
func nonce() -> String { UUID().uuidString.replacingOccurrences(of: "-", with: "").lowercased() + UUID().uuidString.replacingOccurrences(of: "-", with: "").lowercased() }
func aliases() -> [String] {
    for type in [UTType.png, .jpeg, .plainText, .html] {
        let names = (type.tags[.mimeType] ?? []).filter { $0.range(of: "^[a-z0-9!#$%&'+.^_`|~/-]+$", options: .regularExpression) != nil && $0.utf8.count <= 127 }
        if names.count >= 2 && names[0] != names[1] { return Array(names.prefix(2)) }
    }
    return []
}
func crossProcess(_ board: NSPasteboard, _ dir: String, _ record: State) throws {
    if try !owns(board, record) { throw Refusal.ownership }
    struct Representation: Decodable { let MIME: String; let Data: Foundation.Data }
    let expected = try JSONDecoder().decode([Representation].self, from: privateData(dir + "/expected.json", maximum: 32768))
    if expected.count != 5 { throw Refusal.mismatch }
    guard let item = try boardItems(board).first else { throw Refusal.mismatch }
    for representation in expected {
        let type: NSPasteboard.PasteboardType
        switch representation.MIME {
        case "text/plain": type = .string
        case "image/png": type = .png
        case "text/html": type = .html
        default: guard let resolved = UTType(mimeType: representation.MIME) else { throw Refusal.mismatch }; type = .init(resolved.identifier)
        }
        if VellumGuardData(item, type) != representation.Data { throw Refusal.mismatch }
    }
    if try revision(board) != record.revision { throw Refusal.changed }
}
func emit(_ result: [String: Any], _ dir: String? = nil) throws {
    let data = try JSONSerialization.data(withJSONObject: result, options: [.sortedKeys])
    if let dir { try durable(data, dir + "/summary.json") }
    FileHandle.standardOutput.write(data); FileHandle.standardOutput.write(Data([10]))
}
func guardian(_ dir: String, _ binary: String, _ testCase: String) throws -> Bool {
    try privateDirectory(dir)
    if !["normal", "exit-after-owned", "kill-after-owned"].contains(testCase) { throw Refusal.arguments }
    for name in ["original.plist", "state.json", "config.json", "child.pid", "expected.json", "summary.json", "child.stdout", "child.stderr"] {
        if fm.fileExists(atPath: dir + "/" + name) { throw Refusal.privateFile }
    }
    guard let board = VellumGuardGeneral() else { throw Refusal.unreadable }
    let originalRevision = try revision(board)
    let original = try capture(board)
    if try capture(board) != original || revision(board) != originalRevision { throw Refusal.changed }
    _ = try nativeItems(original)
    let encoder = PropertyListEncoder(); encoder.outputFormat = .binary
    let backup = try encoder.encode(original)
    if backup.count > maxBackup { throw Refusal.bounds }
    try durable(backup, dir + "/original.plist")
    if try loadOriginal(dir) != original { throw Refusal.mismatch }
    let identity = nonce()
    try durable(json(State(version: 1, nonce: identity, phase: "original", kind: "", revision: originalRevision, itemCount: 0)), dir + "/state.json")
    let alias = aliases()
    let config: [String: Any] = ["Revision": originalRevision, "Nonce": identity, "StatePath": dir + "/state.json", "ExpectedPath": dir + "/expected.json", "ReadyPath": dir + "/child.pid", "Alias": alias]
    try durable(JSONSerialization.data(withJSONObject: config), dir + "/config.json")
    let child = Process(); child.executableURL = URL(fileURLWithPath: binary)
    child.arguments = ["-fixture-config", dir + "/config.json", "-failure-mode", testCase == "kill-after-owned" ? "hold-after-owned" : testCase == "normal" ? "none" : testCase]
    for name in ["child.stdout", "child.stderr"] { if !fm.createFile(atPath: dir + "/" + name, contents: nil, attributes: [.posixPermissions: 0o600]) { throw Refusal.privateFile } }
    let stdout = try FileHandle(forWritingTo: URL(fileURLWithPath: dir + "/child.stdout")), stderr = try FileHandle(forWritingTo: URL(fileURLWithPath: dir + "/child.stderr"))
    defer { try? stdout.close(); try? stderr.close() }; child.standardOutput = stdout; child.standardError = stderr
    try child.run()
    do { try durable(Data("\(child.processIdentifier)\n".utf8), dir + "/child.pid") }
    catch { child.terminate(); child.waitUntilExit(); throw Refusal.privateFile }
    let deadline = Date().addingTimeInterval(65)
    var forced = false
    while child.isRunning {
        if Date() >= deadline || (testCase == "kill-after-owned" && (try? state(dir).phase) == "owned") {
            forced = true; child.terminate()
            let grace = Date().addingTimeInterval(2)
            while child.isRunning && Date() < grace { Thread.sleep(forTimeInterval: 0.02) }
            if child.isRunning { kill(child.processIdentifier, SIGKILL) }
            break
        }
        Thread.sleep(forTimeInterval: 0.02)
    }
    child.waitUntilExit(); try noLiveChild(dir)
    let current = try state(dir)
    var checked = false
    if testCase == "normal" && child.terminationStatus == 0 && !forced {
        do { try crossProcess(board, dir, current); checked = true } catch { checked = false }
    }
    let currentOwned = try owns(board, current)
    let failureExpected = testCase != "normal" && child.terminationStatus != 0 && current.phase == "owned" && currentOwned && (testCase != "kill-after-owned" || forced)
    try restore(board, dir, original, current); try eraseBackup(dir)
    let pass = checked || failureExpected
    try emit(["result": pass ? "PASS_GENERAL_CLIPBOARD_RESTORED" : "FAIL_TEST_OR_EXTERNAL_READ_RESTORED", "case": testCase, "child_exit": child.terminationStatus, "forced_child_termination": forced, "cross_process_read": checked, "original_raw_items_restored": true, "protected_backup_removed": true, "original_content_logged": false, "alias_available": alias.count == 2, "preferences_changed": false], dir)
    return pass
}
func recover(_ dir: String) throws {
    try noLiveChild(dir)
    let original = try loadOriginal(dir), current = try state(dir)
    guard let board = VellumGuardGeneral() else { throw Refusal.unreadable }
    try restore(board, dir, original, current); try eraseBackup(dir)
    try emit(["result": "PASS_GUARDIAN_RECOVERY_RESTORED", "original_raw_items_restored": true, "protected_backup_removed": true, "original_content_logged": false], dir)
}
func selfTest(_ dir: String) throws {
    try privateDirectory(dir)
    let staleData=Data("private stale guardian journal".utf8)
    let stalePaths=["original.plist","state.json","summary.json"].map { dir + "/" + $0 + ".new" }
    for path in stalePaths { try durable(staleData,path) }
    defer { for path in stalePaths { unlink(path) } }
    guard let board = VellumGuardPrivate() else { throw Refusal.unreadable }; defer { _ = VellumGuardRelease(board) }
    let model = Snapshot(aliases:[],items: [[Flavor(type: NSPasteboard.PasteboardType.string.rawValue, data: Data("guardian original".utf8)), Flavor(type: "org.vellum.guardian.binary", data: Data([0,255,42]))], [Flavor(type: "org.vellum.guardian.empty", data: Data())]])
    try clear(board); if !VellumGuardWrite(board, try nativeItems(model)) { throw Refusal.unreadable }
    let original=try capture(board)
    if original.items != model.items || original.aliases.count != 1 || original.aliases[0].type != legacyStringType.rawValue || original.aliases[0].data != model.items[0][0].data {throw Refusal.mismatch}
    let extra=try boardTypes(board)+[NSPasteboard.PasteboardType("org.vellum.guardian.unrepresented")]
    do { _ = try capture(board,declaredTypes:{extra});throw Refusal.mismatch } catch Refusal.unreadable {}
    let encoder = PropertyListEncoder(); encoder.outputFormat = .binary
    try durable(encoder.encode(original), dir + "/original.plist")
    if try loadOriginal(dir) != original { throw Refusal.mismatch }
    let identity = nonce(), type = try markerType()
    func testItem() throws -> NSPasteboardItem {
        let item = try newItem(); if !VellumGuardSetData(item, Data(identity.utf8), type) { throw Refusal.unreadable }; return item
    }
    try clear(board); if !VellumGuardWrite(board, [try testItem()]) { throw Refusal.unreadable }
    let owned = State(version: 1, nonce: identity, phase: "owned", kind: "rich", revision: try revision(board), itemCount: 1)
    if try !owns(board, owned) { throw Refusal.ownership }
    let admitted = State(version: 1, nonce: identity, phase: "admitted", kind: "rich", revision: try revision(board), itemCount: 1)
    if try owns(board, admitted) { throw Refusal.mismatch }
    let before = try revision(board)
    do { try restore(board, dir, original, admitted); throw Refusal.mismatch } catch Refusal.ownership {}
    if try revision(board) != before { throw Refusal.mismatch }
    try restore(board, dir, original, owned)
    if try capture(board) != original { throw Refusal.mismatch }
    try clear(board); if !VellumGuardWrite(board, [try testItem()]) { throw Refusal.unreadable }
    let prior = State(version: 1, nonce: identity, phase: "owned", kind: "rich", revision: try revision(board), itemCount: 1)
    let foreignItem = try newItem()
    if !VellumGuardSetData(foreignItem, Data("foreign private-board writer".utf8), .string) { throw Refusal.unreadable }
    try clear(board); if !VellumGuardWrite(board, [foreignItem]) { throw Refusal.unreadable }
    let foreign = try capture(board), foreignRevision = try revision(board)
    do { try restore(board, dir, original, prior); throw Refusal.mismatch } catch Refusal.ownership {}
    if try capture(board) != foreign || revision(board) != foreignRevision { throw Refusal.mismatch }
    try clear(board)
    let empty = State(version: 1, nonce: identity, phase: "empty", kind: "empty", revision: try revision(board), itemCount: 0)
    try restore(board, dir, original, empty)
    if try capture(board) != original { throw Refusal.mismatch }; try eraseBackup(dir)
    // Inject actual private-board foreign writes during each journal I/O gap.
    for phase in ["admitted","empty"] {
        try clear(board);if !VellumGuardWrite(board,[try testItem()]) {throw Refusal.unreadable}
        let current=State(version:1,nonce:identity,phase:"owned",kind:"rich",revision:try revision(board),itemCount:1)
        var injected:Snapshot?;var injectedRevision:Int?
        do {
            try restore(board,dir,original,current,interpose:{step,target in
                if step != phase {return}
                let replacement=try newItem()
                if !VellumGuardSetData(replacement,Data("private interposed writer".utf8),.string) {throw Refusal.unreadable}
                try clear(target);if !VellumGuardWrite(target,[replacement]) {throw Refusal.unreadable}
                injected=try capture(target);injectedRevision=try revision(target)
            });throw Refusal.mismatch
        } catch Refusal.ownership {}
        if try injected==nil || capture(board) != injected || revision(board) != injectedRevision {throw Refusal.mismatch}
    }
    try emit(["result": "PASS_PRIVATE_GUARDIAN_SAFETY", "general_pasteboard_access": false, "multiple_raw_items": true, "empty_flavor": true, "binary_flavor": true, "derived_string_alias_preserved":true,"foreign_writer_refused": true, "unknown_admission_refused": true, "known_empty_restore": true, "exact_restore": true,"unrepresented_board_type_refused":true,"foreign_after_admission_refused":true,"foreign_after_empty_refused":true,"stale_journal_ignored":true], dir)
    for path in stalePaths { if try privateData(path,maximum:128) != staleData { throw Refusal.mismatch } }
}

let args = CommandLine.arguments
let lockPath = "/tmp/vellum-clipboard-acceptance-\(geteuid()).lock"
let lockFD = open(lockPath, O_RDWR | O_CREAT | O_NOFOLLOW, 0o600)
var lockInfo = stat()
if lockFD < 0 || fstat(lockFD, &lockInfo) != 0 || lockInfo.st_mode & S_IFMT != S_IFREG || lockInfo.st_uid != geteuid() || lockInfo.st_mode & 0o777 != 0o600 || flock(lockFD, LOCK_EX | LOCK_NB) != 0 {
    FileHandle.standardError.write(Data("clipboard acceptance lock unavailable\n".utf8)); exit(1)
}
defer { flock(lockFD, LOCK_UN); close(lockFD) }
do {
    if args.count == 3 && args[1] == "--self-test" { try selfTest(args[2]) }
    else if args.count == 3 && args[1] == "--recover" { try recover(args[2]) }
    else if args.count == 5 && args[1] == "--run" { if try !guardian(args[2], args[3], args[4]) { exit(1) } }
    else { throw Refusal.arguments }
} catch {
    // No localizedDescription/NSError/exception payload: provider and backup
    // errors can contain original clipboard metadata. Keep those out of logs.
    let code = (error as? Refusal).map { String(describing: $0) } ?? "system"
    FileHandle.standardError.write(Data("clipboard guardian refused or failed (\(code)); protected backup retained if captured\n".utf8)); exit(1)
}
