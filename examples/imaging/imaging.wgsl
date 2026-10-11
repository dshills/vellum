struct Params { source: vec4<f32>, viewport: vec4<f32> }
@group(0) @binding(0) var<uniform> params: Params;
@group(0) @binding(1) var<storage, read_write> pixels: array<vec4<f32>>;
@group(0) @binding(3) var<storage, read> pixels_read: array<vec4<f32>>;
@group(0) @binding(2) var hdr: texture_2d<f32>;

@compute @workgroup_size(8, 8)
fn cs_main(@builtin(global_invocation_id) id: vec3<u32>) {
 let size = vec2<u32>(params.source.yz);
 if (id.x >= size.x || id.y >= size.y) { return; }
 let index = id.y * size.x + id.x;
 if (id.y < size.y / 4u) { pixels[index] = vec4<f32>(0); return; }
 var straight = vec3<f32>(2, 0.5, 0);
 if (id.x >= size.x / 2u) { straight = vec3<f32>(0, 0.25, 3); }
 let alpha = 0.5;
 pixels[index] = vec4<f32>(straight * params.source.x * alpha, alpha);
}

@vertex fn vs_main(@builtin(vertex_index) id: u32) -> @builtin(position) vec4<f32> {
 let vertices = array<vec2<f32>, 3>(vec2<f32>(-1, -1), vec2<f32>(3, -1), vec2<f32>(-1, 3));
 return vec4<f32>(vertices[id], 0, 1);
}
@fragment fn fs_float(@builtin(position) pos: vec4<f32>) -> @location(0) vec4<f32> {
 let p = vec2<u32>(pos.xy);
 return pixels_read[p.y * 256u + p.x];
}
@fragment fn fs_display(@builtin(position) pos: vec4<f32>) -> @location(0) vec4<f32> {
 let uv = pos.xy / params.viewport.xy;
 let size = textureDimensions(hdr);
 let p = min(vec2<u32>(uv * vec2<f32>(size)), size - vec2<u32>(1));
 let value = textureLoad(hdr, vec2<i32>(p), 0);
 if (value.a <= 0) { return vec4<f32>(0); }
 let straight = value.rgb / value.a;
 let mapped = straight / (vec3<f32>(1) + straight);
 return vec4<f32>(mapped * value.a, value.a);
}
