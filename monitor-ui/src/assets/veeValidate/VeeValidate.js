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
      alpha: '只能包含字母字符',
      max(n, e) {
        return '不能超出' + e[0] + '字符'
      },
      min(n, e) {
        return '不能少于' + e[0] + '字符'
      },
      between(n,e){
        return '必须在'+e[0]+','+e[1]+'之间'
      },
      email: '必须为有效邮箱'
    }
  }
})
const config = {
  errorBagName: 'veeErrors', // change if property conflicts.
  fieldsBagName: 'veeFields',
  delay: 0,
  locale: localStorage.getItem('lang') || (navigator.language || navigator.userLanguage === 'zh-CN'
    ? 'zh-CN'
    : 'en-US'),
  messages: null,
  strict: true,
  events: 'blur'
  // events: 'keyup|input|blur'
}

Validator.extend('noEmail', {
  getMessage: () => '格式不正确',
  validate: value => (/^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$/gi.test(value))
})

Validator.extend('isIP', {
  getMessage: () => 'ip格式不正确',
  validate: value => (/((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}/g.test(value))
})

Validator.extend('isNumber', {
  getMessage: () => '输入必须为数字',
  validate: value => (/^-?\d+(\.\d+)?$/.test(value) || /^\d{1,}$/.test(value))
})

export const veeValidateConfig = config
export default VeeValidate
