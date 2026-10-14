# Packed owned span rows

Owned patch 0060 copies a span operation into one exact detached backing buffer.
Rows use full slice expressions with capacity equal to length. Appending cannot
write the next row. Construction storage and spare headers retire as before;
original byte charges, admission order, span order and encodings are unchanged.

The 600-stream original row-layout oracle checks serialized GPU/original bytes,
seeded replay, borrowed construction mutation and appended row isolation. Existing
budget/refused-allocation tests remain in force. New tests are MIT. Preparation
checks all input/output hashes and the patch SHA256.
