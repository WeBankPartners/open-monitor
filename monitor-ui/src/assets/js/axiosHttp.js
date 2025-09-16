import { getToken } from '@/assets/js/cookies.ts'
import axios from 'axios'
import { baseURL_config } from './baseURL'
export default function ajax(options) {
  const ajaxObj = {
    method: options.method,
    baseURL: baseURL_config,
    url: options.url,
    timeout: 30000,
    params: options.params,
    // params: options.params || '',
    headers: {
      'Content-type': 'application/json;charset=UTF-8',
      'X-Auth-Token': getToken() || null,
      Authorization: 'Bearer ' + localStorage.getItem('monitor-accessToken')
    },
    // data: JSON.stringify(options.data || '')
    data: JSON.stringify(options.data),
  }
  // 导出请求时增加响应类型
  if (options.url.endsWith('/export')) {
    ajaxObj.responseType = 'blob'
  }
  return window.request ? window.request(ajaxObj) : axios(ajaxObj)
}
