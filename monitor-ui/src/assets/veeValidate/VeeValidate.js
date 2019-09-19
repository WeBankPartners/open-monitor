import Vue from 'vue'
import VeeValidate, {Validator} from 'vee-validate'
import CN from 'vee-validate/dist/locale/zh_CN'

Validator.localize('zh_CN', CN)
Validator.localize({
  'zh_CN': {
    messages: {
      required: '不能为空',
      numeric: '仅允许数字',
      alpha:'只能包含字母字符',
      max: function (n, e) {
        return '不能超出' + e[0] + '字符'
      },
      min: function (n, e) {
        return '不能少于' + e[0] + '字符'
      },
      between: function(n,e){
        return '必须在'+e[0]+','+e[1]+'之间'
      },
      email: '必须为有效邮箱'
    }
  }
})

const config = {
  errorBagName: 'errors', // change if property conflicts.
  delay: 0,
  locale: 'zh_CN',
  messages: null,
  strict: true,
  events: 'blur'
  // events: 'keyup|input|blur'
}

Validator.extend('noChinese', {
  getMessage: () => '不能包含中文',
  validate: value => {
    return !(/[\u4E00-\u9FA5]|[\uFE30-\uFFA0]/gi.test(value))
  }
})

Validator.extend('noEmail', {
  getMessage: () => '格式不正确',
  validate: value => {
    return (/^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$/gi.test(value))
  }
})

Validator.extend('mobile', {
  getMessage: () => '必须是11位手机号码',
  validate: value => {
    return value.length == 11 && /^((13|14|15|17|18)[0-9]{1}\d{8})$/.test(value)
  }
})

Vue.use(VeeValidate, config)

export default VeeValidate
