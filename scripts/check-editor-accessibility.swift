// Read-only external-process editor AX acceptance. Requires existing caller trust;
// it never prompts for or changes accessibility or input-source preferences.
import ApplicationServices
import Foundation

struct Failure: Error, CustomStringConvertible {
    let description: String
    init(_ text: String) { description = text }
}
func emit(_ value: [String: Any]) {
    if let data = try? JSONSerialization.data(withJSONObject: value, options: [.sortedKeys]),
       let text = String(data: data, encoding: .utf8) { print(text) }
}
func attribute(_ node: AXUIElement, _ name: String) throws -> CFTypeRef {
    var result: CFTypeRef?
    let status = AXUIElementCopyAttributeValue(node, name as CFString, &result)
    guard status == .success, let value = result else {
        throw Failure("attribute \(name): AX error \(status.rawValue)")
    }
    return value
}
func parameter(_ node: AXUIElement, _ name: String, _ value: CFTypeRef) throws -> CFTypeRef {
    var result: CFTypeRef?
    let status = AXUIElementCopyParameterizedAttributeValue(node, name as CFString, value, &result)
    guard status == .success, let returned = result else {
        throw Failure("parameter \(name): AX error \(status.rawValue)")
    }
    return returned
}
func axRange(_ start: Int, _ length: Int) throws -> AXValue {
    var value = CFRange(location: start, length: length)
    guard let result = AXValueCreate(.cfRange, &value) else { throw Failure("range allocation") }
    return result
}
func range(_ value: CFTypeRef) throws -> [Int] {
    guard CFGetTypeID(value) == AXValueGetTypeID() else { throw Failure("expected range AXValue") }
    let ax = unsafeBitCast(value, to: AXValue.self)
    var result = CFRange()
    guard AXValueGetValue(ax, .cfRange, &result) else { throw Failure("expected CFRange") }
    return [result.location, result.length]
}
func rect(_ value: CFTypeRef) throws -> CGRect {
    guard CFGetTypeID(value) == AXValueGetTypeID() else { throw Failure("expected rectangle AXValue") }
    let ax = unsafeBitCast(value, to: AXValue.self)
    var result = CGRect.zero
    guard AXValueGetValue(ax, .cgRect, &result) else { throw Failure("expected CGRect") }
    return result
}
func text(_ value: CFTypeRef) throws -> String {
    guard let result = value as? String else { throw Failure("expected AX string") }
    return result
}
func number(_ value: CFTypeRef) throws -> Int {
    guard let result = value as? NSNumber else { throw Failure("expected AX number") }
    return result.intValue
}
func children(_ node: AXUIElement, _ name: String) -> [AXUIElement] {
    guard let value = try? attribute(node, name), let list = value as? [AXUIElement] else { return [] }
    return list
}
func document(_ app: AXUIElement, until deadline: Date) -> AXUIElement? {
    var queue = [app]
    var visited: [AXUIElement] = []
    var index = 0
    while index < queue.count && index < 512 && Date() < deadline {
        let node = queue[index]; index += 1
        guard !visited.contains(where: { CFEqual($0, node) }) else { continue }
        visited.append(node)
        _ = AXUIElementSetMessagingTimeout(node, 1)
        let role = (try? text(attribute(node, "AXRole"))) ?? ""
        let title = (try? text(attribute(node, "AXTitle"))) ?? ""
        let label = (try? text(attribute(node, "AXDescription"))) ?? ""
        if role == "AXTextArea" && (title == "Document text" || label == "Document text") { return node }
        queue.append(contentsOf: children(node, "AXChildren").prefix(max(0, 512 - queue.count)))
        if CFEqual(node, app) { queue.append(contentsOf: children(node, "AXWindows").prefix(max(0, 512 - queue.count))) }
    }
    return nil
}
func logicalLineCount(_ value: String) -> Int {
    let units = Array(value.utf16)
    var line = 0, i = 0
    while i < units.count {
        if units[i] == 13 {
            line += 1
            if i + 1 < units.count && units[i + 1] == 10 { i += 1 }
        } else if units[i] == 10 { line += 1 }
        i += 1
    }
    return line
}
func check(pid: Int32, expected: String) throws -> [String: Any] {
    let app = AXUIElementCreateApplication(pid)
    let timeoutStatus = AXUIElementSetMessagingTimeout(app, 1)
    guard timeoutStatus == .success else { throw Failure("AX messaging timeout refused") }
    let deadline = Date().addingTimeInterval(12)
    var target: AXUIElement?
    while Date() < deadline {
        if let candidate = document(app, until: deadline), let current = try? text(attribute(candidate, "AXValue")), Array(current.utf8) == Array(expected.utf8) {
            target = candidate; break
        }
        Thread.sleep(forTimeInterval: 0.05)
    }
    guard let node = target else { throw Failure("expected external AXTextArea did not publish") }
    let value = try text(attribute(node, "AXValue"))
    guard Array(value.utf8) == Array(expected.utf8) && value.hasPrefix("A😀e\u{0301}") && value.hasSuffix("\r\n\r\n") else {
        throw Failure("exact UTF-8/CRLF/final-empty-line value mismatch")
    }
    let selected = try range(attribute(node, "AXSelectedTextRange"))
    let selectedText = try text(attribute(node, "AXSelectedText"))
    let characters = try number(attribute(node, "AXNumberOfCharacters"))
    guard selected == [1, 2] && selectedText == "😀" && characters == value.utf16.count else {
        throw Failure("UTF-16 selected range/text/character count mismatch")
    }
    let emoji = try text(parameter(node, "AXStringForRange", axRange(1, 2)))
    guard emoji == "😀" else { throw Failure("external emoji substring mismatch") }
    let composed = try range(parameter(node, "AXRangeForIndex", NSNumber(value: 3)))
    guard composed == [3, 2] else { throw Failure("combining grapheme range mismatch") }
    // Nil/error is refusal; any successful string from a surrogate interior is a failure.
    var split: CFTypeRef?
    let splitError = AXUIElementCopyParameterizedAttributeValue(node, "AXStringForRange" as CFString, try axRange(2, 1), &split)
    guard splitError != .success || split == nil else { throw Failure("split-surrogate range was accepted") }
    let lastLine = logicalLineCount(value)
    let eofLine = try number(parameter(node, "AXLineForIndex", NSNumber(value: characters)))
    let finalRange = try range(parameter(node, "AXRangeForLine", NSNumber(value: lastLine)))
    guard eofLine == lastLine && finalRange == [characters, 0] else { throw Failure("CRLF/final-empty-line range mismatch") }
    var frame = CGRect.zero
    while Date() < deadline {
        if let current = try? rect(parameter(node, "AXBoundsForRange", axRange(1, 2))),
           current.width > 0 && current.height > 0 { frame = current; break }
        Thread.sleep(forTimeInterval: 0.05)
    }
    guard [frame.minX, frame.minY, frame.width, frame.height].allSatisfy({ $0.isFinite }),
          frame.width > 0 && frame.height > 0 else { throw Failure("external emoji range bounds unavailable") }
    var point = CGPoint(x: frame.midX, y: frame.midY)
    guard let axPoint = AXValueCreate(.cgPoint, &point) else { throw Failure("point allocation") }
    let hit = try range(parameter(node, "AXRangeForPosition", axPoint))
    guard hit == [1, 2] else { throw Failure("external screen point did not hit emoji: \(hit)") }
    let hebrew = (value as NSString).range(of: "אבג")
    guard hebrew.location != NSNotFound else { throw Failure("missing bidi corpus") }
    let first = try rect(parameter(node, "AXBoundsForRange", axRange(hebrew.location, 1)))
    let last = try rect(parameter(node, "AXBoundsForRange", axRange(hebrew.location + 2, 1)))
    guard first.width > 0 && last.width > 0 && first.minX > last.minX else { throw Failure("external Hebrew range bounds lack visual RTL order") }
    return ["event": "vellum.editor.external-ax", "result": "pass", "pid": pid,
            "accessibility_trusted": true, "exact_value": true, "characters_utf16": characters,
            "selection_utf16": selected, "selected_text": selectedText,
            "combining_range_utf16": composed, "split_surrogate_error": splitError.rawValue,
            "final_line": eofLine, "final_line_range_utf16": finalRange,
            "emoji_bounds_screen": [frame.minX, frame.minY, frame.width, frame.height],
            "emoji_point_range_utf16": hit, "hebrew_visual_rtl": true,
            "scope": "read-only external AX process; no VoiceOver, real IME or setting changes"]
}
let trusted = AXIsProcessTrusted()
if CommandLine.arguments.count == 2 && CommandLine.arguments[1] == "--trust-only" {
    emit(["event": "vellum.editor.external-ax-trust", "accessibility_trusted": trusted,
          "caller": CommandLine.arguments[0], "prompt": false])
    exit(trusted ? 0 : 2)
}
guard CommandLine.arguments.count == 3, let pid = Int32(CommandLine.arguments[1]), pid > 0 else {
    emit(["result": "failure", "error": "usage: check-editor-accessibility PID EXPECTED_UTF8_FILE"]); exit(1)
}
guard trusted else {
    emit(["result": "unavailable", "accessibility_trusted": false, "prompt": false]); exit(2)
}
do {
    let data = try Data(contentsOf: URL(fileURLWithPath: CommandLine.arguments[2]))
    guard let expected = String(data: data, encoding: .utf8), data.count <= 4 * 1024 * 1024 else { throw Failure("invalid expected UTF-8 fixture") }
    emit(try check(pid: pid, expected: expected))
} catch {
    emit(["event": "vellum.editor.external-ax", "result": "failure", "pid": pid,
          "accessibility_trusted": trusted, "error": String(describing: error)])
    exit(1)
}
