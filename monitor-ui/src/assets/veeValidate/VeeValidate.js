import Vue from 'vue'
import VeeValidate, {Validator} from 'vee-validate'
import zh from 'vee-validate/dist/locale/zh_CN'
import en from 'vee-validate/dist/locale/en'

Validator.localize('en-US', en)
Validator.localize('zh-CN', zh)
Validator.localize({
  'zh-CN': {
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
  locale: localStorage.getItem('lang') || navigator.language || navigator.userLanguage,
  messages: null,
  strict: true,
  events: 'blur'
  // events: 'keyup|input|blur'
}

Validator.extend('noEmail', {
  getMessage: () => '格式不正确',
  validate: value => {
    return (/^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$/gi.test(value))
  }
})


Validator.extend('isIP', {
  getMessage: () => 'ip格式不正确',
  validate: value => {
    return (/((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}/g.test(value))
  }
})

Validator.extend('isNumber', {
  getMessage: () => '输入必须为数字',
  validate: value => {
    return (/^\d{1,}$/.test(value))
  }
})

Vue.use(VeeValidate, config)

export default VeeValidate
