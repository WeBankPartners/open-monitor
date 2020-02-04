import Vue from "vue";
import vuex from "vuex";
Vue.use(vuex);

export default new vuex.Store({
  state: {
    ip: {},
    tableExtendActive: -1, //table组件扩展状态
    recursiveNum: []
  },
  mutations: {
    storeip(state, ip) {
      state.ip = ip
    },
    changeTableExtendActive(state, index) {
      state.tableExtendActive = index
    },
    addRecursiveNum(state, index) {
      state.recursiveNum.push(index)
    },
    removeRecursiveNum(state, index) {
      state.recursiveNum.splice(index, 1)
    },
    emptyRecursiveNum(state) {
      state.recursiveNum = []
    },
  }
});
