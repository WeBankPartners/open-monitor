import isEmpty from 'lodash/isEmpty'
import {validate} from './validate'

const colorSet = ['#487e89', '#395b79', '#153863', '#153250']
// const colorSet = ['#61a0a8', '#2f4554', '#c23531', '#d48265', '#91c7ae', '#749f83', '#ca8622', '#bda29a', '#6e7074', '#546570', '#c4ccd3']
export function generateUuid() {
  return new Promise(resolve => {
    resolve(guid())
  })
}

export function randomColor() {
  const index = Math.floor((Math.random()*colorSet.length))
  return new Promise(resolve => {
    resolve(colorSet[index])
  })
}

function guid() {
  return 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
    const r = Math.random()*16|0, v = c === 'x' ? r : (r&0x3|0x8)
    return v.toString(16)
  })
}

/*
 * Func: 按要求截取字符串
 *
 * @param {String} value (待截取字符串)
 * @param {Int} maxLen (最大长度)
 */
export function interceptParams(value = '', maxLen = 20) {
  if (validate.isEmpty_reset(value)) {
    return ''
  }
  if (value.length > maxLen) {
    return value.substring(0,maxLen) + '...'
  }
  return value
}

export function debounce(fn, delay = 500) {
  let timer = null
  return function () {
    const args = arguments
    if (timer) {
      clearTimeout(timer)
    }
    timer = setTimeout(() => {
      fn.apply(this, [...args])
    }, delay)
  }
}

// 截流函数
export const throttle = (fn, delay) => {
  let timer = null
  const that = this
  return args => {
    if (timer) {
      return
    }
    timer = setTimeout(() => {
      fn.apply(that, args)
      timer = null
    }, delay)
  }
}

// 深拷贝
export const deepClone = obj => {
  const objClone = Array.isArray(obj) ? [] : {}
  if (obj && typeof obj === 'object') {
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        if (obj[key] && typeof obj[key] === 'object') {
          objClone[key] = deepClone(obj[key])
        } else {
          objClone[key] = obj[key]
        }
      }
    }
  }
  return objClone
}

export const showPoptipOnTable = (className='.ivu-poptip-popper') => {
  setTimeout(() => {
    const elements = document.querySelectorAll(className)
    const visibleElements = Array.prototype.filter.call(elements, function (element) {
      return window.getComputedStyle(element).display === 'block'
    })
    if (isEmpty(visibleElements)) {
      return
    }
    const resElement = visibleElements[0]
    const rect = resElement.getBoundingClientRect()
    resElement.style.position = 'fixed'
    resElement.style.top = rect.top + 'px'
    resElement.style.left = rect.left + 'px'
  }, 0)
}

export const isStringFromNumber = str => !isNaN(str) && !isNaN(parseFloat(str))

export const isPositiveNumericString = str => /^\d+$/.test(str) && parseFloat(str) >= 0

export const getRandomColor = () => {
  const letters = '0123456789ABCDEF'
  let color = '#'
  for (let i = 0; i < 6; i++) {
    color += letters[Math.floor(Math.random() * 16)]
  }
  return color
}

export const isSameArray = (arr1, arr2) => {
  if (!Array.isArray(arr1) || !Array.isArray(arr2)) {
    return false
  }
  if (arr1.length !== arr2.length) {
    return false
  }
  const sortedArr1 = arr1.sort()
  const sortedArr2 = arr2.sort()
  for (let i = 0; i < sortedArr1.length; i++) {
    if (sortedArr1[i] !== sortedArr2[i]) {
      return false
    }
  }
  return true
}
