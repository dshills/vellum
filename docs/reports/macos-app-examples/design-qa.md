# macOS app visual QA

Comparison: external `screenshots/source-implementation-comparison.png` places
real Apple references beside separately captured Vellum native windows. The
comparison only crops source guide chrome and resizes images to fit columns;
implementation pixels are not edited. Different states and capture densities
prevent pixel-diff acceptance. This is a qualitative comparison with explicit
remaining gaps.

| App | Reference pixels/state | Vellum pixels/state | Result |
| --- | --- | --- | --- |
| Calculator | 230×408, actual Apple app, prior expression/history | 243×425, 1+2=3, fixed240×420Dp request | Full keypad/no scrolling; missing native mode/history toolbar, different typography/spacing |
| Notes | 2122×1289 guide capture, app crop489,247..1634,1000, Projects/context menu | 1114×749, Notes/meeting fixture | Three columns and centered date work; monospace body/gutter, separate title, missing rich text/toolbar/materials |
| Weather | 1280×720 guide page, app crop232,357..639,615, different location/sidebar state | 1114×770, Cupertino Fahrenheit fixture | Hero centered, cards visible; flat background, different density/native toolbar/materials |
| Reminders | 1280×720 guide page, crop232,296..640,636, Groceries16 | 1013×749, fictional Groceries6 | Full-width title/list/composer and separated counts; rows/sidebar density, exact symbols/native toolbar/materials differ |

The raw native screenshots retain the tool's pixel dimensions; independently
measured backing-scale equality was not established. Display comparison is
normalized to fit, not a claim of equivalent physical or point density. Guide
screenshots have different content and viewport sizes. Apple captures remain
external development evidence and are not bundled or committed.

P1: app-specific native toolbars and adaptive header/safe-area support remain.
P1: Notes lacks public proportional/gutterless rich text presentation.
P2: Tahoe materials, exact symbols, spacing and densities differ across examples.
Functional native keyboard/accessibility checks and software pointer routing are
recorded separately; physical hardware pointer behavior is not qualified here.
No new OS input-source, VoiceOver or global appearance changes were made.

final result: blocked
