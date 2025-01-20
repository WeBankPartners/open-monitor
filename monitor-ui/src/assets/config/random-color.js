const lodash = require('lodash')

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

/**
 * 生成同一色系的颜色数组，严格限制在同一色系
 * @param {string} color - 原始颜色（十六进制格式，如 #ffffff）
 * @param {number} length - 生成的颜色数组长度
 * @returns {string[]} - 返回一个包含颜色的数组
 */
function generateAdjacentColors(color, length) {
  if (length === 1) {
    return [color]
  }
  // 十六进制颜色转 HSL
  const hexToHsl = hex => {
    // eslint-disable-next-line
    hex = hex.replace('#', '')
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

    let h = 0, // Hue
      // eslint-disable-next-line
      s = 0, // Saturation
      // eslint-disable-next-line
      l = (max + min) / 2 // Lightness

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
    // 调整亮度：从 20% 到 80% 分布
    const lightness = 20 + ((i / (length - 1)) * 60)

    // 饱和度保持不变，色调微调（限制在原始色调附近）
    const adjustedHsl = {
      h: baseHsl.h, // 保持色调不变
      s: baseHsl.s, // 饱和度不变
      l: Math.max(0, Math.min(100, lightness)) // 确保亮度在 0%-100% 之间
    }

    colors.push(hslToHex(adjustedHsl.h, adjustedHsl.s, adjustedHsl.l))
  }

  return colors
}

export {
  generateAdjacentColors,
  stringToNumber,
  changeSeriesColor
}
