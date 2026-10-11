# Portable editable Grid review handoff

The historical draft-1 record is a deliberately failing reproduction, not acceptance. Three reviewer test bodies (five cases) expose clipped overscan admission and known cell-version rollback/reuse. The original modified runtime, relevant fixture helpers and exact test bodies are preserved as inert `.go.txt` files. The red Go race command exited 1.

Final draft-3 uses corrected portable runtime. The original bodies are byte-identical. Four additional tests exercise prepared-source CAS/viewport independence, live source drift during candidate reads, version history across unavailable/incarnation states, detached substring retention and generation exhaustion. The isolated `go test -race` command ran them 20 times and exited 0, with unchanged copied inputs. The approval also records the final length-before-UTF8 guard.

The raw review/proof maps retain their original scratch paths as informational provenance. Files in this handoff are the copied artifacts; `handoff.json` verifies their exact UTF-8 byte hashes. Unchanged portable runtime context and auxiliary copied test/dependency files are identified by hashes in the raw source/module maps but their contents are deliberately not duplicated. This bundle is not a complete source archive or whole-dependency proof. Canonical base was ed75cce; root separately qualifies the full integrated portable/UI source.

Scope is portable edit admission, recovery, version/ABA policy, indexed work and payload detachment. MaxEditBytes is per retained edit string; MaxTextBytes remains the separate snapshot payload bound. Native text/IME, GUI, external AX, physical interaction, performance and the independent consumer are excluded. No GUI, Git or Prism was run by this reviewer.
