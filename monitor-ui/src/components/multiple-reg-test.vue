<template>
  <div style="margin: 4px 0px;">
    <Form :label-width="100">
      <FormItem :label="$t('m_log_sample')" style="margin-bottom:12px">
        <Input type="textarea" v-model="textString">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_service_code_regex')" style="margin-bottom:12px">
        <Input type="textarea" v-model="serviceCodeRegex" @on-change="regexChange('serviceCodeRegex')">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_return_code_regex')" style="margin-bottom:12px">
        <Input type="textarea" v-model="returnCodeRegex" @on-change="regexChange('returnCodeRegex')">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_time_regex')" style="margin-bottom:12px">
        <Input type="textarea" v-model="timeConsumingRegex" @on-change="regexChange('timeConsumingRegex')">
        </Input>
      </FormItem>
      <FormItem :label="$t('m_match')">
        <div v-html="regRes" style="word-break: break-all;"></div>
      </FormItem>
      <FormItem style="text-align: right;">
        <Button type="primary" size="small" @click="apiTest">{{$t('m_background_trial')}}</Button>
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
      activeRegex: '',
      regex: '',
      textString: '',
      serviceCodeRegex: '',
      returnCodeRegex: '',
      timeConsumingRegex: '',
      apiRes: ''
    }
  },
  computed: {
    regRes: function () {
      try {
        const reg = new RegExp(this.regex, 'g')
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
    regexChange (tmpRegex) {
      this.activeRegex = tmpRegex
      this.regex = tmpRegex
    },
    apiTest () {
      const params = {
        reg_string: this.regex,
        test_context: this.textString
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/regexp/test/match', params, (responseData) => {
        this.apiRes = responseData
      }, {isNeedloading:false})
    }
  },
  components: {}
}
</script>

<style scoped lang="scss">
</style>