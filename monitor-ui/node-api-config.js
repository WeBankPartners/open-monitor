/* eslint-disable no-template-curly-in-string */
/* eslint-disable no-eval */
/* eslint-disable camelcase */
/*
 *
 * 备注：采用该脚本生成API权限JSON，需注意以下几点：
 * 1. 页面中用到的API必须在server.js中定义
 * 2. 组件中目前识别通过import { *** } from '@/api/server'(.js可带可不带) 这种形式引入的API，其他形式的导入暂不支持。
 * 3. server.js中定义的API必须以export const开头，并且url不能通过变量方式传递，需要写在server.js中
 *    // export const saveBatchExecute = (url, data) => req.post(url, data)，这种方式传递就不行
 * 4. server.js中api的url需要用模板字符串的形式拼接url
 * 5. 自定义写法，需要在相应组件写入自定义组件枚举custom_api_enum
   6. 全局注册组件，通过Vue.component方法注册的，脚本会自动识别兼容
 */

const fs = require('fs')
const path = require('path')
const compiler = require('vue-template-compiler')
const glob = require('glob')
const {find} = require('lodash')

const API_SERVER_PATH = 'src/assets/config/api-center.json' // api入口文件
const GLOBAL_COMPONENT_PATH = './project-config/main.js' // 全局组件入口文件
const WRITE_API_PATH = path.resolve(path.resolve('..'), 'monitor-server/conf') // api-menu-config.json文件写入的路径
// const WRITE_API_PATH = ''
// -------------------------------------------------生成组件对应api的引用关系--------------------------------------------------------------

// 根据import语句返回对应API函数列表
function extractApiImports(content) {
  // 匹配 this.request('POST', this.apiCenter.template.templateSet, params) 这种形式的methods和template.templateSet
  const regex = /this\.request\(\s*['"]([A-Z]+)['"]\s*,\s*this\.apiCenter\.([\w.]+)/g;
  const matches = [...content.matchAll(regex)]
    .map(match => ({ [match[2]]: match[1].toLowerCase() }))
    .filter((obj, index, arr) => 
      arr.findIndex(o => JSON.stringify(o) === JSON.stringify(obj)) === index)
  if (matches) {
    return matches
  }
  return []
}

// 提取 custom_api_enum 内容
function extractCustomApiEnum(content) {
  const regex = /custom_api_enum\s*=\s*(\[.*?\])/gs
  const match = regex.exec(content)
  if (match && match[1]) {
    const apiEnum = eval(match && match[1])
    return apiEnum
  }
  return []
}

function scanFilesAndExtractImports(rootPath) {
  // 使用 glob 匹配所有 .vue 文件
  const files_vue = glob.sync('**/*.vue', { cwd: rootPath })
  // 使用 glob 匹配所有 .js 文件
  const files_js = glob.sync('**/*.js', { cwd: rootPath })
  // 合并 .vue 和 .js 文件列表
  const files = files_vue.concat(files_js)
  const allImports = [] // 存储每个文件及其导入的 API 列表
  files.forEach(file => {
    file = path.resolve(rootPath, file) // 绝对路径
    const content = fs.readFileSync(file, 'utf-8')
    let apiImports = extractApiImports(content)
    let customApi = extractCustomApiEnum(content)
    if (apiImports.length > 0 || customApi.length > 0) {
      allImports.push({
        path: file,
        api: customApi ? apiImports.concat(customApi) : apiImports
      })
    }
  })

  return allImports
}

const rootPath_ = path.resolve(__dirname, 'src')
const allImports = scanFilesAndExtractImports(rootPath_)


// -------------------------------------------生成菜单对应组件path的映射关系---------------------------------------------

// 菜单对应组件入口集合
const menuPathMap = {
  "MONITOR_CUSTOM_DASHBOARD": [
    "src/views/custom-view/view-config-index",
    "src/views/custom-view/chart-list",
    "src/views/custom-view/view-config"
  ],
  "MONITOR_MAIN_DASHBOARD": [
    "src/views/endpoint-view"
  ],
  "MONITOR_ALARM_MANAGEMENT": [
    "src/views/alarm-management",
    "src/views/alarm-history"
  ],
  "MONITOR_ALARM_CONFIG": [
    "src/views/monitor-config/endpoint-management",
    "src/views/monitor-config/group-management",
    "src/views/monitor-config/resource-level",
    "src/views/monitor-config/business-monitor",
    "src/views/monitor-config/log-template",
    "src/views/metric-config/index",
    "src/views/monitor-config/threshold-management",
    "src/views/monitor-config/log-management"
  ],
  "MONITOR_ADMINISTRATOR_CONFIG": [
    "src/views/admin-config/basic/type-config",
    "src/views/admin-config/basic/board-config",
    "src/views/admin-config/basic/metric-config",
    "src/views/admin-config/other/exporter",
    "src/views/admin-config/other/remote-sync"
  ]
}




// 全局注册组件特殊匹配--start
const getGlobalComponetMap = filePath => {
  const fileContent = fs.readFileSync(filePath, 'utf-8')
  // 使用正则表达式匹配Vue.component和import语句
  const vueComponentRegex = /Vue\.component\('([^']*)',\s*([^)]*)\)/g
  const importRegex = /import\s+[^ ]+\s+from\s+['"]([^'"]+)['"]/g

  // 提取Vue.component中的组件名和对应的变量名
  const vueComponents = []
  let vueComponentMatch
  while ((vueComponentMatch = vueComponentRegex.exec(fileContent)) !== null) {
    vueComponents.push({
      componentName: vueComponentMatch[1],
      variableName: vueComponentMatch[2].trim()
    })
  }

  // 提取import语句中的变量名和路径
  const imports = {}
  let importMatch
  while ((importMatch = importRegex.exec(fileContent)) !== null) {
    const importStatement = importMatch[0]
    const importPath = importMatch[1]
    const variableName = importStatement.match(/import\s+([^ ]+)/)[1]
    imports[variableName] = importPath
  }

  // 构建最终的对象
  const result = {}
  vueComponents.forEach(component => {
    if (imports[component.variableName]) {
      imports[component.variableName] = path.resolve(
        path.dirname(GLOBAL_COMPONENT_PATH),
        imports[component.variableName]
      )
      result[component.componentName] = imports[component.variableName]
    }
  })
  return result
}
const globalComponetMap = getGlobalComponetMap(GLOBAL_COMPONENT_PATH)
// 全局注册组件特殊匹配--end

const menuKeysMap = {}
Object.entries(menuPathMap).forEach(([menu, pathArr]) => {
  // 存储同一菜单下的所有组件path
  const allPath = []
  // 带单下对应的组件入口路径集合
  pathArr.forEach(entry_path => {
    const projectRoot = path.resolve(__dirname, '')
    // 绝对路径
    let entryComponentPath = path.resolve(projectRoot, entry_path)
    // 入口文件不是.vue结尾的，自动补全
    if (!entryComponentPath.endsWith('.vue')) {
      entryComponentPath = entryComponentPath + '.vue'
      // 可能组件省略了index.vue结尾，则尝试加上index.vue再查找文件是否存在
      if (!fs.existsSync(entryComponentPath)) {
        entryComponentPath = entryComponentPath.replace(/\.vue$/, '/index.vue')
      }
    }
    // 存储已解析的组件路径，避免重复解析
    const parsedComponents = new Set()

    // 解析组件并提取 key 值
    function parseComponent(filePath) {
      if (parsedComponents.has(filePath)) return
      parsedComponents.add(filePath)
      let scriptContent = ''
      let templateContent = ''
      if (filePath.endsWith('.vue')) {
        const content = fs.readFileSync(filePath, 'utf-8')
        const parsed = compiler.parseComponent(content)
        scriptContent = parsed.script.content
        templateContent = parsed.template.content
      } else if (filePath.endsWith('.js')) {
        scriptContent = fs.readFileSync(filePath, 'utf-8')
        templateContent = fs.readFileSync(filePath, 'utf-8')
      }
      allPath.push(filePath)

      // 通过import语句引用组件
      const importRegex = /import\s+\w+\s+from\s+['"]([^'"]+)['"]/g
      let importMatch
      while ((importMatch = importRegex.exec(scriptContent))) {
        const relativePath = importMatch[1].replace(/@/g, path.resolve(__dirname, 'src')) // 替换 @ 别名
        let absolutePath = path.resolve(path.dirname(filePath), relativePath)
        // 如果import引入的文件不是js或者vue文件结尾, 1.则在后面加上.vue或者.js结尾再尝试查找文件是否存在
        if (!absolutePath.endsWith('.vue') && !absolutePath.endsWith('.js')) {
          absolutePath = absolutePath + '.vue'
          if (!fs.existsSync(absolutePath)) {
            absolutePath = absolutePath.replace(/\.vue$/, '.js')
            // 如果自动添加.vue和.js结尾的文件都不存在，则在后面加上/index.vue再尝试查找文件是否存在
            if (!fs.existsSync(entryComponentPath)) {
              entryComponentPath = entryComponentPath.replace(/\.js$/, '/index.vue')
            }
          }
        }
        if (fs.existsSync(absolutePath)) {
          parseComponent(absolutePath)
        } else {
          // console.error(`文件不存在: ${absolutePath}`)
        }
      }

      // 通过window.component注册的全局组件
      if (globalComponetMap && Object.keys(globalComponetMap).length > 0) {
        const globalComponetNames = Object.keys(globalComponetMap)
        const usedComponentNames = globalComponetNames.filter(name => templateContent.indexOf(`<${name}`) > -1) || []
        usedComponentNames.forEach(name => {
          const componentPath = globalComponetMap[name]
          if (fs.existsSync(componentPath)) {
            parseComponent(componentPath)
          } else {
            console.error(`文件不存在: ${componentPath}`)
          }
        })
      }
    }

    // 从入口组件开始解析
    parseComponent(entryComponentPath)
  })
  menuKeysMap[menu] = allPath
})

// -----------------------------------------------生成菜单对应api集合映射关系-----------------------------------------------

// 得到菜单和对应api的集合
const apiMap = {}
allImports.forEach(item => {
  apiMap[item.path] = item.api
    /* 数据结构如下
    'D:\\webankCode\\wecube-platform\\wecube-portal\\src\\pages\\admin\\base-migration\\import\\history.vue': [
      'getBaseMigrationImportList',
      'getBaseMigrationImportQuery',
      'updateImportStatus',
      [{
        "url": "/platform/v1/process/definitions/export",
        "method": "post"
      }]
      ]
    */
  // }
})



// 遍历第一个JSON对象，将对应的API合并到一个数组中，并去重
for (const [key, values] of Object.entries(menuKeysMap)) {
  menuKeysMap[key] = Array.from(
    values.reduce((acc, path) => {
      if (apiMap[path]) {
        apiMap[path].forEach(api => acc.add(api)) // 使用Set去重
      }
      return acc
    }, new Set())
  )
}
// const jsonString_ = JSON.stringify(menuKeysMap, null, 2)
// fs.writeFileSync(path.join(__dirname, 'node-api-tree.json'), jsonString_, 'utf-8')

// -------------------------------------------------将api转换成key、method、url组成的对象-----------------------------------------

// 读取代码文件
const filePath = path.join(__dirname, API_SERVER_PATH)
const code = fs.readFileSync(filePath, 'utf-8')

let apiConfigArr = []
const getApiConfigArr = () => {
  apiConfigArr = []
  function flattenObject(obj, parentKey = '', result = {}) {
    for (const key in obj) {
      const currentValue = obj[key];
      const newKey = parentKey ? `${parentKey}.${key}` : key;

      if (typeof currentValue === 'object' && !Array.isArray(currentValue)) {
        // 如果是对象，递归处理子属性
        flattenObject(currentValue, newKey, result);
      } else {
        // 如果是字符串或其他类型，直接赋值
        result[newKey] = currentValue;
      }
    }
    return result;
  }

  let allApiObj = flattenObject(JSON.parse(code))


  Object.keys(allApiObj).forEach(key => {
    const url = allApiObj[key]
    apiConfigArr.push({
      key,
      url: url.replace(/\?.*/, '')
    })
  })
}

// const getApiConfigArr = () => {
//   // 将server.js文件切割成api组成的数组结构
//   const exportRegex = /export\s+const\s+([\s\S]+?)(?=\nexport\s+const\s+|$)/g
//   const apiArray = []
//   let match

//   while ((match = exportRegex.exec(code)) !== null) {
//     const apiDefinition = match[1].trim()
//     apiArray.push(apiDefinition)
//   }
//   apiArray.forEach(apiStr => {
//     // 使用正则表达式匹配 key、method 和 url
//     const keyMatch = apiStr.match(/(\w+)\s*=/)
//     // const methodUrlMatch = apiStr.match(/req\.(\w+)\(['"`]([^'"`]+)['"`]/)
//     // 优化正则，可以识别req.post()里面内容换行的情况
//     const methodUrlMatch = apiStr.match(/req\.(\w+)\(\s*['"`]([^?'"`]*?)(?:\?[^'"`]*?)?['"`]/s)
//     if (keyMatch && methodUrlMatch) {
//       const key = keyMatch[1] // 函数名称
//       const method = methodUrlMatch[1].toLowerCase() // 请求方法名
//       const url = methodUrlMatch[2].replace(/\${(.*?)}/g, '${$1}') // 接口 URL
//       apiConfigArr.push({
//         key,
//         url: '/wecmdb/api/v1' + url,
//         method
//       })
//     }
//   })
//   // 将结果写入 JSON 文件
//   // const outputFilePath = path.join(__dirname, 'api_config.json');
//   // fs.writeFileSync(outputFilePath, JSON.stringify(apiConfigArr, null, 2));
//   // console.log(`API 配置已生成并保存到 ${outputFilePath}`);
// }
getApiConfigArr()



// --------------------------------------------------------合并数据生成最终结果--------------------------------------------------------



const newMenuKeysMap = {}
Object.keys(menuKeysMap).forEach(key => {
  newMenuKeysMap[key] = []
  menuKeysMap[key].forEach(apiName => {
    const item = find(apiConfigArr, {
      key: Object.keys(apiName)[0]
    })
    newMenuKeysMap[key].push({
      ...item,
      method: Object.values(apiName)[0]
    })


    // apiConfigArr.forEach(item => {
    //   if (item.key === apiName) {
    //     newMenuKeysMap[key].push(item)
    //   }
    // })
    // if (Array.isArray(apiName) && apiName.length > 0) {
    //   // 组件内部暴露的自定义api custom_api_enum
    //   newMenuKeysMap[key].push(...apiName)
    // }
  })
})

let finalResult = []
Object.keys(newMenuKeysMap).forEach(key => {
  const menuUrlsObj = {
    menu: key,
    urls: newMenuKeysMap[key]
  }
  finalResult.push(menuUrlsObj)
})

const menuApiMapPath = path.join(WRITE_API_PATH, 'menu-api-map.json')
fs.writeFileSync(menuApiMapPath, JSON.stringify(finalResult, null, 2))
   