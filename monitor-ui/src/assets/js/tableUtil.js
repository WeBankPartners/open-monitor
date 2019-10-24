/*
* @author: 冯经宇
* @CreateDate: 2018-01-16
* @update: 2018-12-25
* @version: V0.1.1
* @describe:
* 该插件旨在将table类型公用js抽象于此
*
*/

/*
 * Func: 初始化table内容
 *
 * @param {Object} that (页面当前对象)
 * @param {String} method
 * @param {String} url
 * @param {Object} pageConfig (页面整天参数配置)
 */
const initTable = (that, method, url, pageConfig) =>{
  // 将搜索组件和分页组件中条件合并
  let filters = pageConfig.researchConfig?that.$validate.isEmptyReturn_JSON(that.$validate.deepCopy(pageConfig.researchConfig.filters)):{}
  let params = that.$validate.deepCopy(pageConfig.pagination)
  for (let k in filters) {
    params[k] = filters[k]
  }
  let requestParams = adapterParamsForTabledata(that, params)
  return that.$httpRequestEntrance.httpRequestEntrance(method, url, requestParams, (responseData) => {
    that.pageConfig.pagination.total = responseData.num
    that.pageConfig.pagination.current = parseInt(params.page)
    let res = that.$validate.isEmpty_reset(responseData.data) ? [] : responseData.data
    that.pageConfig.table.tableData = res
    return that.pageConfig.pagination
  })
}

/*
 * Func: 剔除无需传入后台的请求字段
 *
 * @param {Object} oriParams (待处理数据)
 *
 * return {Object} requestParams (处理后数据)
 */
const adapterParamsForTabledata = (that, oriParams) =>{
  let requestParams = that.$validate.deepCopy(oriParams)
  let deleteparams = ['current', 'pageSize', 'total', 'pageCount']
  for (let k in requestParams) {
    for (let i = 0; i < deleteparams.length; i++){
      if (k === deleteparams[i]) {
        delete requestParams[k]
      }
    }
  }
  return requestParams
}

/*
 * Func: table编辑功能时，将row中数据赋值给AddParams
 *
 * @param {Object} that (vue实例对象)
 * @param {Object} AddParams (待渲染对象)
 * @param {Object} that (row数据)
 */
const manageEditParams = (AddParams, rowparams) =>{
  for (let key in AddParams) {
    AddParams[key] = rowparams[key]
  }
  return AddParams
}

/*
 * Func: 初始化详情页中table内容
 *
 * @param {Object} that (页面当前对象)
 * @param {String} method
 * @param {String} url
 * @param {Object} pageConfig (页面整天参数配置)
 */
const initDetailTable = (_this, indexx) =>{
  _this.detailPageConfig.detailConfig[indexx].table.tableData = []
  let methods = _this.detailPageConfig.detailConfig[indexx].pagination.getData.methods
  let url = _this.detailPageConfig.detailConfig[indexx].pagination.getData.url
  // 将搜索组件和分页组件中条件合并
  let filters = _this.detailPageConfig.detailConfig[indexx].researchConfig ? _this.$validate.isEmptyReturn_JSON(_this.$validate.deepCopy(_this.detailPageConfig.detailConfig[indexx].researchConfig.filters)) : null
  let params = Object.assign({}, _this.detailPageConfig.detailConfig[indexx].pagination)
  // let params = _this.$validate.deepCopy(_this.detailPageConfig.detailConfig[indexx].pagination)
  for (let k in filters) {
    params[k] = filters[k]
  }
  params.__offset = (params.current-1)*params.__limit
  // 剔除无需传入后台的字段
  let deleteparams = ['current', 'pageSize', 'total', 'pageCount','getData']
  for (let k in params) {
    for (let i = 0; i < deleteparams.length; i++){
      if (k === deleteparams[i]) {
        delete params[k]
      }
    }
  }
  _this.$httpRequestEntrance.httpRequestEntrance(methods, url, params, (responseData) => {
    _this.detailPageConfig.detailConfig[indexx].table.tableData = responseData[_this.detailPageConfig.detailConfig[indexx].pagination.getData.data]
    _this.detailPageConfig.detailConfig[indexx].pagination.total = responseData[_this.detailPageConfig.detailConfig[indexx].pagination.getData.count]
  })
}

export const tableUtil = {
  initTable,
  manageEditParams,
  initDetailTable
}
