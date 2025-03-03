/*
* @version: V0.1.1
* @describe:
* 统一http请求入口，统一处理http请求响应
*
 */
// import router from '../../../project-config/router'
import httpRequest from '@/assets/js/axiosHttp'
import { setLocalstorage, clearLocalstorage } from '@/assets/js/localStorage.js'
import Vue from 'vue'
import axios from 'axios'
import $ from 'jquery'
import loadingContent from '@/assets/static-file/loading.json'
import {Message} from 'view-design'

let loadingCount = 0
// 页面loading配置
export const loading = {
  start: () => {
    const htmlLevel1 ='<div id="loadingContainer" class="loadingContainer" style="width: 100%;height: 100%;position: fixed;bottom: 0;text-align: center;opacity: 0.5;z-index:9000">'
    const hmtlLevel2=`<img src=${loadingContent.content} class="loadingImg" style="display: inline-block;width: 4rem; height: 4rem;position: absolute;top: 50%; left: 50%; margin-top: -62px; margin-left: -62px;"></div>`
    if ($('#loadingContainer')) {
      $('#loadingContainer').remove()
      if (window.request) {
        // within wecube platform
        $('#wecube_app>div>div.content-container').append(htmlLevel1 + hmtlLevel2)
      } else {
        // standalone
        $('#app>div>div.content').append(htmlLevel1 + hmtlLevel2)
      }
    }
  },
  end: () => {
    if ($('#loadingContainer') && loadingCount<1) {
      $('#loadingContainer').remove()
    }
  }
}
// 错误消息提醒统一组件
function errorMessage(content) {
  Message.error({
    content,
    duration: 5,
    closable: true
  })
}

const throwError = res => {
  Vue.prototype.$Notice.warning({
    title: 'Error',
    desc: (res.data && + res.data.status + '<br/> ' + res.data.message) || 'error',
    duration: 10
  })
}
/*
 * Func: http统一处理
 *
 * @param 你懂得
 */
function httpRequestEntrance(method, url, data, callback, customHttpConfig, errCallback) {
  // 处理接口http请求个性化配置
  const config = mergeObj(customHttpConfig)
  let finalCallBack = errCallback
  if (config.isNeedloading) {
    loadingCount++
    loading.start()
  }
  const option = {
    method,
    url
  }
  if (method.toUpperCase() === 'GET' || method.toUpperCase() === 'DELETE') {
    option.params = data
  } else {
    option.data = data
  }
  option.timeout = config.timeout

  return new Promise(resolve => {
    httpRequest(option).then(response => {
      if (config.isNeedloading) {
        setTimeout(() => {
          loadingCount--
          loading.end()
        },0)
      }
      if (window.request) {
        if (response.status === 'OK') {
          resolve(response.data)
          return callback ? callback(response.data, response.message, response.data) : null
        }
      } else {
        if (response.status < 400) {
          if (response.data.status === 'OK') {
            resolve(response.data.data)
            return callback ? callback(response.data.data ,response.data.message, response.data) : null
          }
          errorMessage(response.data.message)
        }
      }

    })
      .catch(function (err) {
        if (config.isNeedloading) {
          setTimeout(() => {
            loadingCount--
            loading.end()
          },0)
        }
        const { response } = err
        if (response.status === 401 && err.config.url !== '/auth/v1/api/login') {
          const refreshToken = localStorage.getItem('monitor-refreshToken')
          if (refreshToken.length > 0) {
            const refreshRequest = axios.get('/auth/v1/api/token', {
              headers: {
                Authorization: 'Bearer ' + refreshToken
              }
            })
            return refreshRequest.then(
              resRefresh => {
                setLocalstorage(resRefresh.data.data)
                // replace token with new one and replay request
                err.config.headers.Authorization = 'Bearer ' + localStorage.getItem('monitor-accessToken')
                const retryRequest = axios(err.config)
                return retryRequest.then(
                  res => {
                    if (res.status === 200) {
                    // do request success again
                      if (res.data.status === 'ERROR') {
                        const errorMes = Array.isArray(res.data.data)
                          ? res.data.data.map(_ => _.message || _.errorMessage).join('<br/>')
                          : res.data.message
                        Vue.prototype.$Notice.warning({
                          title: 'Error',
                          desc: errorMes,
                          duration: 10
                        })
                      }
                      return res.data instanceof Array ? res.data : { ...res.data }
                    }

                    return {
                      data: throwError(res)
                    }

                  },
                  err => {
                    const { response } = err
                    return new Promise(resolve => {
                      resolve({
                        data: throwError(response)
                      })
                    })
                  }
                )
              },
              // eslint-disable-next-line handle-callback-err
              () => {
                clearLocalstorage()
                window.location.href = window.location.origin + window.location.pathname + '#/login'
                return {
                  data: {} // throwError(errRefresh.response)
                }
              }
            )
          }
          window.location.href = window.location.origin + window.location.pathname + '#/login'
          if (response.config.url === '/auth/v1/api/login') {
            Vue.prototype.$Notice.warning({
              title: 'Error',
              desc: response.data.message || '401',
              duration: 10
            })
          }
          // throwInfo(response)
          return response

        }
        // errorMessage(error.response.data.message)
        err.response&&err.response.data&&errorMessage(err.response.data.message)
        // if (!window.request && error.response && error.response.status === 401) {
        //   router.push({path: '/login'})
        // }
        if (typeof customHttpConfig === 'function') {
          finalCallBack = customHttpConfig
        }
        if (typeof finalCallBack === 'function' && err.response && err.response.data) {
          return finalCallBack(new Error(err.response && err.response.data ? err.response.data.message : ''))
        }
      })
  })
}

/*
 *Func: 处理接口http请求个性化配置
 * @param {Object} config (个性化配置)
 * @param {Object} res (个性化配置与公共配置合并结果)
 *
 */

function mergeObj(config) {
  const httpConfig = {
    isNeedloading: true,
    timeout: 30000
  }
  const res = Object.assign(httpConfig, config)
  return res
}
export default {
  httpRequestEntrance
}
