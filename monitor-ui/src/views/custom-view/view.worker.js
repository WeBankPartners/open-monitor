// 导入模块函数
import { initPanalsWebWorker } from './workerContent'

self.onmessage = e => {
  const result = initPanalsWebWorker(e.data)
  self.postMessage(result)
}
