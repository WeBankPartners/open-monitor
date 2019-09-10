import Vue from 'vue'
import vuex from 'vuex'
Vue.use(vuex)

export default new vuex.Store({
  state:{
    ip: {}
  },
  mutations:{
    storeip (state, ip) {
      state.ip = ip    
    }
  }
})