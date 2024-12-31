const lodash = require('lodash')
const hexToHSL = H => {
  const r = parseInt(H.substring(1, 3), 16) / 255
  const g = parseInt(H.substring(3, 5), 16) / 255
  const b = parseInt(H.substring(5, 7), 16) / 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const delta = max - min
  const l = (max + min) / 2
  let h = 0
  let s = 0
  if (delta !== 0) {
    s = l < 0.5 ? delta / (max + min) : delta / (2 - max - min)
    switch (max) {
      case r: h = (g - b) / delta + (g < b ? 6 : 0); break
      case g: h = (b - r) / delta + 2; break
      case b: h = (r - g) / delta + 4; break
    }
    h /= 6
  }
  return [h * 360, s * 100, l * 100]
}

const hslToHex = (h, s, l) => {
  const c = (1 - Math.abs(2 * l - 1)) * s
  const x = c * (1 - Math.abs((h / 60) % 2 - 1))
  const m = l - c / 2
  let r = 0
  let g = 0
  let b = 0
  if (0 <= h && h < 60) {
    r = c; g = x; b = 0
  } else if (60 <= h && h < 120) {
    r = x; g = c; b = 0
  } else if (120 <= h && h < 180) {
    r = 0; g = c; b = x
  } else if (180 <= h && h < 240) {
    r = 0; g = x; b = c
  } else if (240 <= h && h < 300) {
    r = x; g = 0; b = c
  } else if (300 <= h && h < 360) {
    r = c; g = 0; b = x
  }
  r = Math.round((r + m) * 255).toString(16)
  g = Math.round((g + m) * 255).toString(16)
  b = Math.round((b + m) * 255).toString(16)
  if (r.length === 1) {
    r = '0' + r
  }
  if (g.length === 1) {
    g = '0' + g
  }
  if (b.length === 1) {
    b = '0' + b
  }
  return '#' + r + g + b
}

/**
 * 生成同一色系的颜色数组
 * @param {string} color - 原始颜色（十六进制格式，如 #ffffff）
 * @param {number} length - 生成的颜色数组长度
 * @returns {string[]} - 返回一个包含颜色的数组
 */
function generateAdjacentColors22(color, length) {
  if (length === 1) {
    return [color]
  }
  // 十六进制颜色转 RGB
  const hexToRgb = hex => {
    const tempHex = hex.replace('#', '')
    const bigint = parseInt(tempHex, 16)
    return {
      r: (bigint >> 16) & 255,
      g: (bigint >> 8) & 255,
      b: bigint & 255,
    }
  }

  // RGB 转十六进制颜色
  const rgbToHex = (r, g, b) => {
    const toHex = value => value.toString(16).padStart(2, '0')
    return `#${toHex(r)}${toHex(g)}${toHex(b)}`
  }
  // 调整颜色亮度
  const adjustBrightness = function (rgb, factor) {
    return {
      r: Math.max(0, Math.min(255, Math.round(rgb.r * factor))),
      g: Math.max(0, Math.min(255, Math.round(rgb.g * factor))),
      b: Math.max(0, Math.min(255, Math.round(rgb.b * factor))),
    }
  }
  // 转换原始颜色为 RGB
  const baseRgb = hexToRgb(color)
  // 生成颜色数组
  const colors = []
  for (let i = 0; i < length; i++) {
    // 生成亮度因子（在 0.7 到 1.7 之间随机浮动）
    const factor = 0.6 + (i / (length - 1)) * 1 // 让亮度从暗到亮分布
    const adjustedRgb = adjustBrightness(baseRgb, factor)
    colors.push(rgbToHex(adjustedRgb.r, adjustedRgb.g, adjustedRgb.b))
  }
  return colors
}

/**
 * 生成同一色系的颜色数组，区分度更高
 * @param {string} color - 原始颜色（十六进制格式，如 #ffffff）
 * @param {number} length - 生成的颜色数组长度
 * @returns {string[]} - 返回一个包含颜色的数组
 */
function generateAdjacentColors(color, length) {
  if (length === 1) {
    return [color]
  }
  // 十六进制颜色转 HSL
  const hexToHsl = temphex => {
    const hex = temphex.replace('#', '')
    const bigint = parseInt(hex, 16)
    const r = (bigint >> 16) & 255
    const g = (bigint >> 8) & 255
    const b = bigint & 255
    const rNorm = r / 255,
      gNorm = g / 255,
      bNorm = b / 255

    const max = Math.max(rNorm, gNorm, bNorm),
      min = Math.min(rNorm, gNorm, bNorm)
    const delta = max - min

    let h = 0 // Hue
    let s = 0 // Saturation
    const l = (max + min) / 2 // Lightness

    if (delta !== 0) {
      h = max === rNorm
        ? ((gNorm - bNorm) / delta) % 6
        : max === gNorm
          ? (bNorm - rNorm) / delta + 2
          : (rNorm - gNorm) / delta + 4

      h = Math.round(h * 60)
      if (h < 0) {h += 360}

      s = delta / (1 - Math.abs(2 * l - 1))
    }

    return {
      h, // 色调
      s: +(s * 100).toFixed(1), // 饱和度（百分比）
      l: +(l * 100).toFixed(1) // 亮度（百分比）
    }
  }

  // HSL 转十六进制颜色
  const hslToHex = (h, s, l) => {
    // eslint-disable-next-line
    s /= 100
    // eslint-disable-next-line
    l /= 100

    const c = (1 - Math.abs(2 * l - 1)) * s,
      x = c * (1 - Math.abs((h / 60) % 2 - 1)),
      m = l - c / 2

    let r = 0, g = 0, b = 0
    if (h >= 0 && h < 60) {
      r = c
      g = x
      b = 0
    } else if (h >= 60 && h < 120) {
      r = x
      g = c
      b = 0
    } else if (h >= 120 && h < 180) {
      r = 0
      g = c
      b = x
    } else if (h >= 180 && h < 240) {
      r = 0
      g = x
      b = c
    } else if (h >= 240 && h < 300) {
      r = x
      g = 0
      b = c
    } else if (h >= 300 && h < 360) {
      r = c
      g = 0
      b = x
    }

    r = Math.round((r + m) * 255).toString(16)
      .padStart(2, '0')
    g = Math.round((g + m) * 255).toString(16)
      .padStart(2, '0')
    b = Math.round((b + m) * 255).toString(16)
      .padStart(2, '0')

    return `#${r}${g}${b}`
  }

  // 转换原始颜色为 HSL
  const baseHsl = hexToHsl(color)

  // 生成颜色数组
  const colors = []
  for (let i = 0; i < length; i++) {
    // 动态调整亮度和色调
    const lightness = baseHsl.l + (i - length / 2) * (100 / length) * 0.6
    const hue = (baseHsl.h + i * 5) % 360 // 稍微调整色调

    const adjustedHsl = {
      h: hue,
      s: baseHsl.s,
      l: Math.max(5, Math.min(95, lightness)) // 确保亮度在 5%-95% 之间
    }

    colors.push(hslToHex(adjustedHsl.h, adjustedHsl.s, adjustedHsl.l))
  }

  return colors
}

const generateAdjacentColors11 = (hexColor, count, degree) => {
  const [h, s, l] = hexToHSL(hexColor)
  const adjacentColors = []
  let tempH = h
  for (let i = 0; i < count; i++) {
    tempH = (tempH + degree) % 360 // 根据传入的度数进行增加
    adjacentColors.push(hslToHex(tempH, s / 100, l / 100))
  }
  return adjacentColors
}
const stringToNumber = (str, min = 1, max = 99) => {
  if (!str) {return Math.floor(Math.random() * (max - min + 1)) + min}
  const hash = Array.from(str).reduce((acc, char) => acc + char.charCodeAt(0), 0)
  const reversed = parseInt((((hash % (max - min + 1)) + min) + '').split('').reverse()
    .join(''), 10)
  const minOutput = 10
  const maxOutput = 90
  return ((reversed - min) / (reversed - max)) * (maxOutput - minOutput) + minOutput
}

const changeSeriesColor = (series = [], color = '') => {
  if (series.length === 0 || !color) {return}
  if (series.length === 1) {
    series[0].color = color
  } else {
    const len = series.length
    const colorArr = generateAdjacentColors(color, len, 5)
    const seriesNameList = series.map(item => item.seriesName)
    seriesNameList.sort()
    for (let i=0; i<seriesNameList.length; i++) {
      const item = lodash.find(series, {
        seriesName: seriesNameList[i]
      })
      item.color = colorArr[i]
    }
  }
  return series
}

export {
  generateAdjacentColors,
  stringToNumber,
  changeSeriesColor,
  generateAdjacentColors11, // 之前的写法，现在已经不用了
  generateAdjacentColors22
}
