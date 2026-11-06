// The app icon: swift scripts/make_icon.swift resources/icon.png
import Foundation
import CoreGraphics
import ImageIO
import UniformTypeIdentifiers

let size = 1024
let out = CommandLine.arguments[1]
let cs = CGColorSpace(name: CGColorSpace.sRGB)!
let ctx = CGContext(data: nil, width: size, height: size, bitsPerComponent: 8, bytesPerRow: 0, space: cs, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
func rgb(_ h: UInt32, _ a: CGFloat = 1) -> CGColor {
  CGColor(colorSpace: cs, components: [CGFloat((h>>16)&255)/255, CGFloat((h>>8)&255)/255, CGFloat(h&255)/255, a])!
}
// macOS icon grid: an 824 square at 100, corners of 185.
let tile = CGRect(x: 100, y: 100, width: 824, height: 824)
let tilePath = CGPath(roundedRect: tile, cornerWidth: 185, cornerHeight: 185, transform: nil)
// A soft drop shadow, as macOS icons have.
ctx.saveGState()
ctx.setShadow(offset: CGSize(width: 0, height: -10), blur: 28, color: rgb(0x000000, 0.28))
ctx.addPath(tilePath); ctx.setFillColor(rgb(0x0F766E)); ctx.fillPath()
ctx.restoreGState()
// The tile: the accent, a little lighter at the top.
ctx.saveGState()
ctx.addPath(tilePath); ctx.clip()
let grad = CGGradient(colorsSpace: cs, colors: [rgb(0x0B5C55), rgb(0x16897F)] as CFArray, locations: [0, 1])!
ctx.drawLinearGradient(grad, start: CGPoint(x: 0, y: 100), end: CGPoint(x: 0, y: 924), options: [])
ctx.restoreGState()
// The eight-pointed star of the Mushaf's ornaments, as on the share cards.
let cx: CGFloat = 512, cy: CGFloat = 512
let R: CGFloat = 300
let stroke: CGFloat = 30
let half = R * 0.8 / sqrt(2)
func point(_ r: CGFloat, _ deg: CGFloat) -> CGPoint {
  let a = deg * .pi / 180
  return CGPoint(x: cx + r * cos(a), y: cy + r * sin(a))
}
let outer = half * sqrt(2), inner = half / cos(.pi / 8)
let star = CGMutablePath()
for k in 0..<16 {
  let p = point(k % 2 == 1 ? inner : outer, CGFloat(k) * 22.5 + 22.5/2*0)
  if k == 0 { star.move(to: p) } else { star.addLine(to: p) }
}
star.closeSubpath()
let cream = rgb(0xFBF5E6)
ctx.setStrokeColor(cream); ctx.setLineWidth(stroke); ctx.setLineJoin(.round)
ctx.addPath(star); ctx.strokePath()
ctx.addEllipse(in: CGRect(x: cx-R, y: cy-R, width: 2*R, height: 2*R)); ctx.strokePath()
let dot = half * 0.42
ctx.setFillColor(cream)
ctx.addEllipse(in: CGRect(x: cx-dot/2, y: cy-dot/2, width: dot, height: dot)); ctx.fillPath()
let img = ctx.makeImage()!
let dest = CGImageDestinationCreateWithURL(URL(fileURLWithPath: out) as CFURL, UTType.png.identifier as CFString, 1, nil)!
CGImageDestinationAddImage(dest, img, nil)
CGImageDestinationFinalize(dest)
