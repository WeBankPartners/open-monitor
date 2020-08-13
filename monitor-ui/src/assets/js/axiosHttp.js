import axios from 'axios'
import {baseURL_config} from './baseURL'
import { getToken } from '@/assets/js/cookies.ts'
const baseURL = `${baseURL_config}/api/v1`
export default function ajax (options) {
  const ajaxObj = {
    method: options.method,
    baseURL: baseURL,
    url: options.url,
    timeout: 30000,
    params: options.params,
    // params: options.params || '',
    headers: {
      'Content-type': 'application/json;charset=UTF-8',
      'X-Auth-Token': getToken() || null
    },
    // data: JSON.stringify(options.data || '')
    data: JSON.stringify(options.data)
  }
  // 导出请求时增加响应类型
  console.log('monitor')
  // if (options.url.endsWith('/export')) {
  //   ajaxObj.responseType = 'blob'
  // }
  return window.request ? window.request(ajaxObj) : axios(ajaxObj)
}
