struct ImageUniform { transform: vec4<f32>, options: vec4<f32> };
@group(0) @binding(0) var<uniform> image_view: ImageUniform;
@group(0) @binding(1) var image_texture: texture_2d<f32>;
struct Output { @builtin(position) position: vec4<f32>, @location(0) uv: vec2<f32> };
@vertex fn vs_image(@builtin(vertex_index) index: u32) -> Output {
 let corners = array<vec2<f32>,6>(vec2(0.,0.),vec2(1.,0.),vec2(0.,1.),vec2(0.,1.),vec2(1.,0.),vec2(1.,1.));
 var out: Output;
 out.uv = corners[index];
 out.position = vec4(out.uv*image_view.transform.xy+image_view.transform.zw,0.5,1.);
 return out;
}
fn premultiplied(point: vec2<i32>, size: vec2<i32>) -> vec4<f32> {
 let pixel = textureLoad(image_texture,clamp(point,vec2<i32>(0),size-vec2<i32>(1)),0);
 return vec4(pixel.rgb*pixel.a,pixel.a);
}
@fragment fn fs_image(in: Output) -> @location(0) vec4<f32> {
 let size = vec2<i32>(textureDimensions(image_texture));
 if image_view.options.x < 0.5 {
  return premultiplied(vec2<i32>(floor(in.uv*vec2<f32>(size))),size);
 }
 let location = in.uv*vec2<f32>(size)-vec2<f32>(0.5);
 let base = vec2<i32>(floor(location));
 let weight = fract(location);
 let top = mix(premultiplied(base,size),premultiplied(base+vec2<i32>(1,0),size),weight.x);
 let bottom = mix(premultiplied(base+vec2<i32>(0,1),size),premultiplied(base+vec2<i32>(1,1),size),weight.x);
 return mix(top,bottom,weight.y);
}
