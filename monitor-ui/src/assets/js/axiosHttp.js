import axios from 'axios'
import {baseURL_config} from './baseURL'
// import {cookies} from '@/common/cookieUtils'
// const baseURL =  require('../../config/APIs').API
const baseURL = `${baseURL_config}/api/v1`
export default function ajax (options) {
  const ajaxObj = {
    method: options.method,
    baseURL: baseURL,
    url: options.url,
    timeout: 30000,
    params: encodeURI(JSON.stringify(options.params)),
    // params: options.params || '',
    headers: {
      'Content-type': 'application/json;charset=UTF-8',
      // 'X-Auth-Token': cookies.getAuthorization() || null
    },
    // data: JSON.stringify(options.data || '')
    data: encodeURI(JSON.stringify(options.data))
  }
  // 导出请求时增加响应类型
  if (options.url.includes('/export/')) {
    ajaxObj.responseType = 'blob'
  }
  return window.request ? window.request(ajaxObj) : axios(ajaxObj)
}
