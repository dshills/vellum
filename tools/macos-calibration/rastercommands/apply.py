#!/usr/bin/env python3
"""Bounded private metadata probe only; no render/source object references retained."""
from pathlib import Path
import shutil
root=Path('/private/tmp/vellum-raster-command-core');sdk=Path('/private/tmp/vellum-doclinks-sdk-bd3cbc9/retained SDK/vellum-v0.1.0-alpha.2/drivers/cogent/build/upstream');code=Path(__file__).resolve().parent
p=root/'paint/renderers/rasterx/renderer.go';original=(sdk/'paint/renderers/rasterx/renderer.go').read_bytes();assert p.read_bytes()==original
s=original.decode();old='func (rs *Renderer) Render(r render.Render) render.Renderer {';assert s.count(old)==1;p.write_text(s.replace(old,old+'\n\trs.vellumCaptureCommands(r)'))
shutil.copyfile(code/'testdata/commands.go',root/'paint/renderers/rasterx/vellum_commands.go')
