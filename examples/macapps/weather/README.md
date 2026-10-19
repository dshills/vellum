# Weather

A native Go consumer of Vellum's public UI descriptions, with a location sidebar,
current conditions, hourly forecast, ten-day forecast and conditions cards. City
selection, saved-location search and Fahrenheit/Celsius presentation work. The
three city forecasts are bounded, deterministic demonstration data, not current
weather. Nothing accesses location, personal accounts, WeatherKit, network data,
notifications or persistence.

Visual reference: [Apple's macOS Tahoe 26 Weather guide](https://support.apple.com/guide/weather-mac/welcome/6.0/mac/26),
read and visually captured in the hidden in-app browser on 19 October 2026. It
shows the centered current conditions, hourly panel, ten-day rows and blue card
hierarchy. Apple reference images are not copied into the shipped application.

The example uses rounded, bordered forecast panels, bundled condition symbols,
and passive low/high temperature bars on one common ten-day scale. All public
controls have accessible names. Search is limited to the three saved sample
cities and retains at most 128 Unicode characters; filtering preserves the
selected forecast, including when there are no matching results. Unit changes
always convert the original Fahrenheit fixture, so repeated toggling has no
rounding drift. The sample city times and forecast date remain fixed.

The city-name button selects its card; the rest of the card remains display-only.
The public UI currently has no image-backdrop/overlay capability, so the sky is
flat blue rather than an animated or photographic sky. This example does not
replicate Apple maps, weather alerts, animated conditions, translucent macOS
materials or detailed forecast drill-downs. Wind and visibility units stay in
the fixed fixture units while the temperature selector changes temperatures.
Native visual acceptance is separate from model tests and is coordinated by the
main development agent. No pixel-identical or native-tested claim is made by
this package alone.

`New`, `Model.View` and `Theme` form the consumer interface; no Core or driver
types are imported. The shared macapps runner owns the model, window and event
loop. Model tests exercise city identity and callback routing, search bounds and
empty results, both temperature presentations, accessible validated views for
every city, the ten-day scale and semantic artwork. These checks provide no
physical-pointer or native-rendering evidence.

The unified-chrome demo reserves 52 device-independent pixels for the native
header measured on macOS Tahoe 26.6.2. Sidebar controls and the forecast scroll
viewport start below that band while their blue backgrounds continue across the
full window surface. The reserve belongs to this tested demo configuration; it
is not an adaptive framework safe-area API or a guarantee for other chrome,
toolbar, accessibility, or OS configurations.
