import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import '@/assets/css/local.bootstrap.css'
import 'bootstrap/dist/js/bootstrap.min.js'
import 'font-awesome/css/font-awesome.css'
import './plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import {tableUtil} from '@/assets/js/tableUtil.js'
import {validate} from '@/assets/js/validate.js'
import VeeValidate from '@/assets/veeValidate/VeeValidate'
import apiCenter from '@/assets/config/api-center.json'

Vue.prototype.$httpRequestEntrance = httpRequestEntrance
Vue.prototype.JQ = jquery
Vue.prototype.$validate = validate
Vue.prototype.$tableUtil = tableUtil
Vue.prototype.apiCenter = apiCenter

import PageTable from '@/components/components/table-page/page'
import ModalComponent from '@/components/components/modal'
Vue.component('PageTable', PageTable)
Vue.component('ModalComponent', ModalComponent)
Vue.use(VeeValidate)

Vue.config.productionTip = false


new Vue({
  render: h => h(App),
  store,
  router
}).$mount('#app')
