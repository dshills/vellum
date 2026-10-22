#!/usr/bin/env python3
"""Apply private bounded command/decision observer to an exact fresh SDK copy."""
from pathlib import Path
import shutil
root=Path('/private/tmp/vellum-raster-native-core');sdk=Path('/private/tmp/vellum-doclinks-sdk-bd3cbc9/retained SDK/vellum-v0.1.0-alpha.2/drivers/cogent/build/upstream');code=Path(__file__).resolve().parent
p=root/'paint/renderers/rasterx/renderer.go';s=(sdk/'paint/renderers/rasterx/renderer.go').read_text();assert p.read_text()==s
def replace(old,new):
 global s
 assert s.count(old)==1,(old,s.count(old));s=s.replace(old,new)
replace('type Renderer struct {','type Renderer struct {\n vellumCommandFrame, vellumCommandDepth int')
replace('func (rs *Renderer) Render(r render.Render) render.Renderer {','func (rs *Renderer) Render(r render.Render) render.Renderer {\n rs.vellumCommandFrame=rs.vellumCaptureCommands(r)\n rs.vellumCommandDepth=0\n defer func(){rs.vellumCommandFrame=-1;rs.vellumCommandDepth=0}()')
replace('\trs.scrollStats.DescriptorBytes = rs.scrollPlanBytes\n\treturn rs','\trs.scrollStats.DescriptorBytes = rs.scrollPlanBytes\n vellumFinishCommands(rs.vellumCommandFrame)\n\treturn rs')
replace('func (rs *Renderer) renderItems(r render.Render, clip image.Rectangle) {','func (rs *Renderer) renderItems(r render.Render, clip image.Rectangle) {\n rs.vellumCommandDepth++\n defer func(){rs.vellumCommandDepth--}()')
replace('\t\t\trs.RenderPath(x)','\t\t\trs.RenderPath(x)\n rs.vellumCommandDecision(i,"render-path")')
replace('coveredUniformDraw(x, next) {\n\t\t\t\t\tcontinue','coveredUniformDraw(x, next) {\n rs.vellumCommandDecision(i,"covered-next")\n\t\t\t\t\tcontinue')
replace('if rs.drawUncoveredUniform(r, i, x) {\n\t\t\t\t\tcontinue','if rs.drawUncoveredUniform(r, i, x) {\n rs.vellumCommandDecision(i,"cropped-uniform")\n\t\t\t\t\tcontinue')
replace('\t\t\t\tx.Render(rs.image)','\t\t\t\tx.Render(rs.image)\n rs.vellumCommandDecision(i,"render-image")')
p.write_text(s)
shutil.copyfile(code/'testdata/native_commands.go',root/'paint/renderers/rasterx/vellum_commands.go')
