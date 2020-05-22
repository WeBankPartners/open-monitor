/*
* @author: pobu168
* @CreateDate: 2019-09-04
* @version: V0.1.1
* @describe:
* 统一http请求入口，统一处理http请求响应
*
 */
// import router from '@/router'
import httpRequest from '@/assets/js/axiosHttp'
import $ from 'jquery'
// import {Message} from 'view-design'

let loadingCount = 0
// 页面loading配置
export const loading = {
  start: ()=>{
    let htmlLevel1 ='<div id="loadingContainer" class="loadingContainer" style="width: 100%;height: 100%;position: fixed;bottom: 0;text-align: center;opacity: 0.5;z-index:9000">'
    let hmtlLevel2='<img src='+ require('@/assets/img/loading.gif') +' class="loadingImg" style="display: inline-block;width: 4rem; height: 4rem;position: absolute;top: 50%; left: 50%; margin-top: -62px; margin-left: -62px;"></div>'
    if ($('#loadingContainer')) {
      $('#loadingContainer').remove()
      $('body').append(htmlLevel1 + hmtlLevel2)
    }
  },
  end: () => {
    if($('#loadingContainer') && loadingCount<1) {
      $('#loadingContainer').remove()
    }
  }
}
// 错误消息提醒统一组件
// function errorMessage(content) {
//   Message.error({
//     content: content,
//     duration: 5,
//     closable: true
//   })
// }

/*
 * Func: http统一处理
 *
 * @param 你懂得
 */
function httpRequestEntrance (method, url, data, callback, customHttpConfig) {
  // 处理接口http请求个性化配置
  let config = mergeObj(customHttpConfig)
  if (config.isNeedloading) {
    loadingCount++
    loading.start()
  }
  let option = {method: method, url: url}
  if (method.toUpperCase() === 'GET' || method.toUpperCase() === 'DELETE' ) {
    option.params = data
  } else {
    option.data = data
  }
  option.timeout = config.timeout
  return httpRequest(option).then(response => {
    if (config.isNeedloading) {
      setTimeout(() => {
        loadingCount--
        loading.end()
      },0)
    }
    if (window.request) {
      return callback(response,response.msg)
    } else {
      if (response.status < 400 && callback !== undefined) {
        return callback(response.data,response.msg)
      }
    }
  }).catch(function () {
    if (config.isNeedloading) {
      setTimeout(() => {
        loadingCount--
        loading.end()
      },0)
    }
  })
}

/*
 *Func: 处理接口http请求个性化配置
 * @param {Object} config (个性化配置)
 * @param {Object} res (个性化配置与公共配置合并结果)
 *
 */

function mergeObj(config) {
  let httpConfig = {
    isNeedloading: true,
    timeout: 30000
  }
  let res = Object.assign(httpConfig, config)
  return res
}
export default {
  httpRequestEntrance
}
