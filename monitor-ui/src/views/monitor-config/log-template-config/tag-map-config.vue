<template>
  <div>
    <Modal v-model="showModel" :title="$t('m_tag_mapping')" :mask-closable="false" :closable="false" :width="500">
      <div>
        <Row>
          <Col span="7">{{ $t('m_match_type') }}</Col>
          <Col span="7">{{ $t('m_source_value') }}</Col>
          <Col span="8">{{ $t('m_match_value') }}</Col>
        </Row>
        <Row v-for="(item, itemIndex) in tagMap" :key="itemIndex" style="margin:6px 0">
          <Col span="7">
          <Select v-model="item.regulative" style="width:90%">
            <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
            <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
          </Select>
          </Col>
          <Col span="7">
          <Input v-model.trim="item.source_value" style="width:90%"></Input>
          </Col>
          <Col span="8">
          <Input v-model.trim="item.target_value" style="width:90%"></Input>
          </Col>
          <Col span="2">
          <Button
            type="error"
            ghost
            @click="deleteItem(itemIndex)"
            size="small"
            style="vertical-align: sub;cursor: pointer"
            icon="md-trash"
          ></Button>
          </Col>
        </Row>
        <div style="text-align: right;margin-right: 9px;cursor: pointer">
          <Button type="success" ghost @click="addItem" size="small" icon="md-add"></Button>
        </div>
      </div>
      <template slot='footer'>
        <Button @click="showModel = false">{{ $t('m_button_cancel') }}</Button>
        <Button @click="okSelect" type="primary">{{ $t('m_button_confirm') }}</Button>
      </template>
    </Modal>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      showModel: false,
      tagMap: []
    }
  },
  methods: {
    loadPage(tagMap) {
      this.tagMap = tagMap
      this.showModel = true
    },
    dataValidateFirst() {
      let res = true
      const infoSet = new Set()
      this.tagMap.forEach(item => {
        if (infoSet.has(item.label)) {
          res = false
        } else {
          infoSet.add(item.label)
        }
      })
      if (!res) {
        this.$Message.warning(this.$t('tw_duplicate_data'))
      }
      return res
    },
    okSelect() {
      // const isCanBeSave = this.dataValidateFirst()
      const isCanBeSave = true
      if (!isCanBeSave) {
        return
      }
      this.$emit(
        'setTagMap',
        this.tagMap.filter(t => t.source_value !== '' && t.target_value !== '')
      )
      this.showModel = false
    },
    addItem() {
      this.tagMap.push(
        {
          regulative: 0, // 匹配类型： 0 是非正则，1是正则
          source_value: '', // 源值
          target_value: '', // 映射值
          value_type: '', // 值类型： success 成功，fail 失败
        }
      )
    },
    deleteItem(itemIndex) {
      this.tagMap.splice(itemIndex, 1)
    }
  }
}
</script>
