<template>
  <div style="margin: 4px 0px;">
    <Form :label-width="100">
      <FormItem :label="$t('m_log_sample')" style="margin-bottom:12px">
        <Input type="textarea" v-model="textString" :rows="4">
        </Input>
      </FormItem>
      <FormItem :label="$t('tableKey.regular')">
        <Input type="textarea" v-model="regx">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_match')">
        <div v-html="regRes" style="word-break: break-all;"></div>
      </FormItem>
      <FormItem style="text-align: right;">
        <Button type="primary" size="small" :disabled="textString==='' || regx===''" @click="apiTest">{{$t('m_background_trial')}}</Button>
      </FormItem>
      <FormItem>
        {{apiRes}}
      </FormItem>
    </Form>
  </div>
</template>

<script>
export default {
  name: '',
  data () {
    return {
      regx: '3245',
      textString: 'eqrwgtsf',
      // regRes: '',
      apiRes: ''
    }
  },
  computed: {
    regRes: function () {
      try {
        const reg = new RegExp(this.regx, 'g')
        let execRes = this.textString.match(reg)
        if (execRes && execRes.length > 0) {
          return this.textString.replace(reg, "<span style='color:red'>" + execRes[0] + '</span>')
        }
        return ''
      } catch (err) {
        return ''
      }
    }
  },
  methods: {
    cancelReg () {
      this.regx = ''
      this.textString = ''
      this.$emit('cancelReg')
    },
    saveReg () {
      this.$emit('updateReg', this.regx)
      this.cancelReg()
    },
    apiTest () {
      const params = {
        reg_string: this.regx,
        test_context: this.textString
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/regexp/test/match', params, (responseData) => {
        this.apiRes = responseData
        this.$emit('setRegResult', ['aaa', 'bbb'], this.textString, this.regx)
      }, {isNeedloading:false})
    }
  },
  components: {}
}
</script>

<style scoped lang="scss">
</style>