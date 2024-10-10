<template>
  <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
    <Form :label-width="80">
      <FormItem :label="$t('m_regular')">
        <Input type="textarea" v-model="regx">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_text')">
        <Input type="textarea" v-model="textString">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_match')">
        <div v-html="regRes" style="word-break: break-all;"></div>
      </FormItem>
      <FormItem>
        <Button type="primary" size="small" @click="apiTest">{{$t('m_background_trial')}}</Button>
      </FormItem>
      <FormItem>
        {{apiRes}}
      </FormItem>
      <FormItem style="text-align: right">
        <Button size="small" @click="cancelReg">{{$t('m_button_cancel')}}</Button>
        <Button size="small" @click="saveReg" type="info">{{$t('m_button_save')}}</Button>
      </FormItem>
    </Form>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      regx: '',
      textString: '',
      // regRes: '',
      apiRes: ''
    }
  },
  computed: {
    regRes() {
      try {
        const reg = new RegExp(this.regx, 'g')
        const execRes = this.textString.match(reg)
        if (execRes && execRes.length > 0) {
          return this.textString.replace(execRes[0], '<span style=\'color:red\'>' + execRes[0] + '</span>')
        }
        return ''
      } catch (err) {
        return ''
      }
    }
  },
  methods: {
    cancelReg() {
      this.regx = ''
      this.textString = ''
      this.$emit('cancelReg')
    },
    saveReg() {
      this.$emit('updateReg', this.regx)
      this.cancelReg()
    },
    apiTest() {
      const params = {
        reg_string: this.regx,
        test_context: this.textString
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/regexp/test/match', params, responseData => {
        this.apiRes = responseData
      }, {isNeedloading: false})
    }
  },
  components: {}
}
</script>

<style scoped lang="scss">
</style>
