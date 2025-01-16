<template>
  <Modal v-model="isSelectModalShow"
         :title="$t('m_line_display_modification')"
         :mask-closable="false"
         :width="1000"
         @on-visible-change="onLineSelectChangeCancel"
  >
    <div v-if="isSelectModalShow">
      <Form :label-width="80" v-if="!isEmpty(lineSelectModalData[chartId]) && Object.keys(lineSelectModalData[chartId]).length > 0">
        <FormItem :label="$t('m_line_search')">
          <Input v-model.trim="lineNameSearch" clearable style="width: 300px" />
          <Checkbox v-model="islineSelectAll"
                    style="margin-left: 10px"
                    @on-change="onLineSelectAllChange"
          >
            <span>{{$t('m_select_all')}}</span>
          </Checkbox>
        </FormItem>
        <FormItem :label="$t('m_show_line')">
          <Row v-if="allShowLineName.length" style="min-height: 200px; max-height: 400px;overflow-y: auto;">
            <Col span="12" v-for="(seriesName, index) in allShowLineName" :key="index">
            <Checkbox v-model="lineSelectModalData[chartId][seriesName]" @on-change="onSingleLineSelectChange">
              <Tooltip :content="seriesName" transfer :max-width='400'>
                <div class="ellipsis-text-style">{{ seriesName }}</div>
              </Tooltip>
            </Checkbox>
            </Col>
          </Row>
          <span v-else style="margin-left: 10px">{{$t('m_noData')}}</span>
        </FormItem>
      </Form>
      <span v-else>
        {{ $t('m_noData') }}
      </span>
    </div>
    <template slot='footer'>
      <Button @click="onLineSelectChangeCancel(false)">{{ $t('m_button_cancel') }}</Button>
      <Button @click="onLineSelectChange" type="primary">{{ $t('m_button_confirm') }}</Button>
    </template>
  </Modal>
</template>

<script>
import {cloneDeep, isEmpty} from 'lodash'
export default {
  name: '',
  props: {
    isLineSelectModalShow: {
      type: Boolean,
      default: false
    },
    chartId: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      isSelectModalShow: false,
      lineSelectModalData: {},
      lineNameSearch: '',
      isEmpty
    }
  },
  watch: {
    isLineSelectModalShow(val) {
      if (val) {
        if (!isEmpty(window['view-config-selected-line-data'])) {
          this.lineSelectModalData = cloneDeep(window['view-config-selected-line-data'])
        }
        if (!isEmpty(this.lineSelectModalData[this.chartId]) && Object.values(this.lineSelectModalData[this.chartId]).every(item => item === true)) {
          this.islineSelectAll = true
        } else {
          this.islineSelectAll = false
        }
        this.lineNameSearch = ''
        this.isSelectModalShow = true
      } else {
        this.isSelectModalShow = false
      }
    }
  },
  computed: {
    allShowLineName() {
      if (Object.keys(this.lineSelectModalData[this.chartId]).length > 0) {
        return Object.keys(this.lineSelectModalData[this.chartId]).filter(item => item.indexOf(this.lineNameSearch) !== -1)
      }
      return []
    }
  },
  methods: {
    onLineSelectChangeCancel(isShow = false) {
      if (!isShow && this.isLineSelectModalShow) {
        this.lineSelectModalData = cloneDeep(window['view-config-selected-line-data'])
        this.$emit('modalClose')
      }
    },
    onLineSelectAllChange(val) {
      for (const line in this.lineSelectModalData[this.chartId]) {
        this.lineSelectModalData[this.chartId][line] = val
      }
    },
    onLineSelectChange() {
      window['view-config-selected-line-data'] = cloneDeep(this.lineSelectModalData)
      this.onLineSelectChangeCancel(false)
    },
    onSingleLineSelectChange() {
      if (!isEmpty(this.lineSelectModalData[this.chartId]) && Object.values(this.lineSelectModalData[this.chartId]).every(item => item === true)) {
        this.islineSelectAll = true
      } else {
        this.islineSelectAll = false
      }
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.ellipsis-text-style {
  width: 350px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: bottom;
}
</style>
