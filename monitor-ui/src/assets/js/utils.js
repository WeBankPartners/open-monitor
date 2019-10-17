import {validate} from './validate'
export function generateUuid () {
    return new Promise((resolve)=>{
        resolve(guid())
    })
}

function guid() {
    return 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random()*16|0, v = c == 'x' ? r : (r&0x3|0x8);
        return v.toString(16);
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