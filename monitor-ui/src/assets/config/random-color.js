const  hexToHSL = (H) => {
  let r = parseInt(H.substring(1, 3), 16) / 255,
      g = parseInt(H.substring(3, 5), 16) / 255,
      b = parseInt(H.substring(5, 7), 16) / 255;
  let max = Math.max(r, g, b),
      min = Math.min(r, g, b);
  let delta = max - min,
      l = (max + min) / 2,
      h = 0,
      s = 0;
  if (delta !== 0) {
      s = l < 0.5 ? delta / (max + min) : delta / (2 - max - min);
      switch (max) {
          case r: h = (g - b) / delta + (g < b ? 6 : 0); break;
          case g: h = (b - r) / delta + 2; break;
          case b: h = (r - g) / delta + 4; break;
      }
      h /= 6;
  }
  return [h * 360, s * 100, l * 100];
}

const hslToHex = (h, s, l) => {
  let c = (1 - Math.abs(2 * l - 1)) * s,
      x = c * (1 - Math.abs((h / 60) % 2 - 1)),
      m = l - c / 2,
      r = 0,
      g = 0,
      b = 0;
  if (0 <= h && h < 60) {
      r = c; g = x; b = 0;
  } else if (60 <= h && h < 120) {
      r = x; g = c; b = 0;
  } else if (120 <= h && h < 180) {
      r = 0; g = c; b = x;
  } else if (180 <= h && h < 240) {
      r = 0; g = x; b = c;
  } else if (240 <= h && h < 300) {
      r = x; g = 0; b = c;
  } else if (300 <= h && h < 360) {
      r = c; g = 0; b = x;
  }
  r = Math.round((r + m) * 255).toString(16);
  g = Math.round((g + m) * 255).toString(16);
  b = Math.round((b + m) * 255).toString(16);
  if (r.length == 1) r = "0" + r;
  if (g.length == 1) g = "0" + g;
  if (b.length == 1) b = "0" + b;
  return "#" + r + g + b;
}

const generateAdjacentColors = (hexColor, count, degree) => {
  let [h, s, l] = hexToHSL(hexColor);
  let adjacentColors = [];

  for (let i = 0; i < count; i++) {
      h = (h + degree) % 360; // 根据传入的度数进行增加
      adjacentColors.push(hslToHex(h, s / 100, l / 100));
  }

  return adjacentColors;
}

export {
  generateAdjacentColors
}
// 示例
// let colors = generateAdjacentColors("#ff5733", 3, 10);
// console.log(colors);