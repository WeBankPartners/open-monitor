/*
* @author: 冯经宇
* @CreateDate: 2017-12-20
* @version: V0.0.1
* @describe:
* 该插件旨在提供公共表单验证插件,功能不断完善中......
*
 */
import $ from 'jquery'

/*
 * Func:单一验证字段是否为空并提示功能
 *
 * @param {String} className (提示信息装载位置的class名称)
 * @param {String} val (待校验字段值)
 * @return {Boolean}
 */
const isEmptyAndWarn = (className, val) => {
  const resClassName = '.' + className
  $(resClassName).empty()
  if (val === '' || val === null || val === undefined || JSON.stringify(val) === '{}' || JSON.stringify(val) === '[]') {
    $(resClassName).append('字段不能为空！')
    return true
  }
  return false
}

/*
 * Func:批量验证字段是否为空并提示功能
 *
 * params: {host_name_w: "111", host_w: "222", os_family_w: "333", checkedEnv_w: "444"}
 * key: 代表提示信息装载位置的class名称(*_w),其中*表示字段名，_w为特殊标记
 * value: 代表待校验字段的值
 */
const isEmptyAndWarn_JSON = params => {
  let res = true
  let key
  for (key in params){
    if (isEmptyAndWarn(key, params[key])) {
      res = false
      break
    }
  }
  return res
}

/*
 * Func:批量验证表单搜索条件
 *
 * params: {host_name: "111", host: "222", os_family: "333"}
 * key:
 * value: 代表待校验字段的值
 */
const isEmptyReturn_JSON = param => {
  const params = Object.assign({},param)
  let key
  for (key in params){
    if (!isEmpty(params[key])) {
      delete params[key]
    }
  }
  return params
}
// 对象内是否有值没填
const isEmptyInObj = params => {
  let key
  for (key in params){
    if (!isEmpty(params[key])) {
      return true
    }
  }
  return false
}

/*
 * Func:单一验证字段是否
 *
 * val: 代表待校验字段的值
 *
 * return: boolean
 */
const isEmpty = val => {
  if (val === '' || val === null || val === undefined || JSON.stringify(val) === '{}' || JSON.stringify(val) === '[]') {
    return false
  }
  return true
}

/*
 * Func:单一验证字段是否
 *
 * val: 代表待校验字段的值
 *
 * return: boolean
 */
const isEmpty_reset = val => {
  if (val === '' || val === null || val === undefined || JSON.stringify(val) === '{}' || JSON.stringify(val) === '[]') {
    return true
  }
  return false
}

/*
 * Func:判断object/json 是否为空
 *
 * e: object/json
 *
 * return: boolean
 */
const isEmptyObject = e => {
  let t
  for (t in e) {
    return !1
  }
  return !0
}

/*
 * Func: 判断字符串是否包含中文
 *
 * @param {String} str (待效验字符串)
 *
 * @return {Boolean} true: 包含; false: 不包含
 */
const isContainChina = str => {
  const patrn=/[\u4E00-\u9FA5]|[\uFE30-\uFFA0]/gi
  if (!patrn.exec(str)){
    return false
  }
  return true
}

/*
 *Func: 清空JSON中字段值，解决编辑和新增时页面缓存
 *
 */
const emptyJson = j => {
  let k
  for (k in j) {
    j[k] = null
  }
  return j
}

/**
 * 深拷贝js对象
 * @param obj
 * @returns {{}}
 *  Created by fengjingyu on 2018/4/11.
 */
const deepCopy = obj => {
  let newO = {}
  if (obj instanceof Array) {
    newO = []
  }
  for (const key in obj) {
    let val = obj[key]
    if (val instanceof Date) {
      val = val.format('yyyy-MM-dd hh:mm:ss')
    }
    // newO[key] = typeof val === 'object' ? arguments.callee(val) : val
    newO[key] = val
  }
  return newO
}

// 格式化时间
Date.prototype.format = function (fmt) {
  let tempFmt = fmt
  const o = {
    'M+': this.getMonth()+1, // 月份
    'd+': this.getDate(), // 日
    'h+': this.getHours(), // 小时
    'm+': this.getMinutes(), // 分
    's+': this.getSeconds(), // 秒
    'q+': Math.floor((this.getMonth()+3)/3), // 季度
    S: this.getMilliseconds() // 毫秒
  }
  if (/(y+)/.test(tempFmt)) {
    tempFmt=tempFmt.replace(RegExp.$1, (this.getFullYear()+'').substr(4 - RegExp.$1.length))
  }
  for (const k in o) {
    if (new RegExp('('+ k +')').test(fmt)){
      tempFmt = tempFmt.replace(RegExp.$1, (RegExp.$1.length===1) ? (o[k]) : (('00'+ o[k]).substr((''+ o[k]).length)))
    }
  }
  return tempFmt
}

const modal_confirm_custom = (leader,title,fun) => {
  const tip = title.length>0 ? '<p style="font-size:15px;transform: translateY(-3px);">'+ title +'</p>' :'确定删除吗?'
  leader.$Modal.confirm({
    content: tip,
    onOk: () => {
      fun()
    },
    onCancel: () => {
      leader.$root.$store.commit('changeFlag',true)
    }
  })
}

const valueFromExpression = (item, expression, default_value='') => {
  const expr=expression
  let item_tmp = item
  const attrs=expr.split('.')
  let n=0
  for (n in attrs){
    if (isEmpty(item_tmp) && attrs[n] in item_tmp){
      item_tmp = item_tmp[attrs[n]]
    } else {
      return default_value
    }
  }
  return item_tmp
}

/*
 *Func: 通过select选中值返回选中对象
 * @param {String} val (选中值)
 * @param {String} key (匹配字段)
 * @param {Object} options (待匹配对象集合)
 *
 */
const returnObjByValue = (val, key, options) => {
  for (const no in options) {
    if (options[no][key] === val) {
      return options[no]
    }
  }
}

// 对象数组转对象
const switchFormat=(arr,key,value) => {
  const obj = {}
  arr.map(item => {
    obj[item[key]] = item[value]
    return item
  })
  return obj
}
/*
 *Func: 精确格式化数字
 * @param {Number} num (待格式化数字)
 * @param {Number} decimalDigit (需要保留小数位)
 * @param {Number} result (返回值)
 *
 */
// const formatNumber = (num, decimalDigit) => {
//   let res = ''
//   let splitNum = num.toString().split('.')
//   if (splitNum.length === 1) {
//     let resNum = splitNum.concat(['.'], new Array(decimalDigit).fill(0))
//     res = resNum.join('')
//   } else {
//     if (splitNum[1].length > decimalDigit) {
//       splitNum[1] = splitNum[1].slice(0, decimalDigit)
//     }
//     res = splitNum.join('.')
//   }
//   return Number(res)
// }

export const validate = {
  isEmpty,
  isEmpty_reset, // 为空效验新方法，计划替代isEmpty
  isEmptyAndWarn,
  isEmptyAndWarn_JSON,
  isEmptyReturn_JSON,
  isEmptyObject,
  isEmptyInObj,
  emptyJson, // 清空JSON对象中的值
  deepCopy, // JSON对象深拷贝
  isContainChina, // 是否包含中文字符
  modal_confirm_custom,
  valueFromExpression, // 多级表达式取值
  returnObjByValue, // 根据选中值返回选中对象
  // formatNumber, // 精确格式化数字
  switchFormat// 对象数组转对象
}
