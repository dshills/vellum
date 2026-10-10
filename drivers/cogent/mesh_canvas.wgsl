// Positions are world coordinates; colors are premultiplied linear RGBA.
@group(0) @binding(0) var<uniform> view: vec4<f32>;
struct VertexOutput {
 @builtin(position) position: vec4<f32>,
 @location(0) color: vec4<f32>,
};
@vertex fn vs_main(@location(0) position: vec2<f32>, @location(1) color: vec4<f32>) -> VertexOutput {
 var out: VertexOutput;
 out.position = vec4<f32>((position + view.zw) * view.xy, 0.5, 1.0);
 out.color = color;
 return out;
}
@fragment fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> { return in.color; }
