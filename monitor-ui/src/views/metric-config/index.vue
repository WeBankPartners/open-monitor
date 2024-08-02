<template>
  <div class="monitor-metric-config">
    <Tabs v-model="activeName">
      <!--通用对象-->
      <TabPane v-if="from === 'admin'" :label="$t('m_basic_type')" name="1">
      </TabPane>
      <!--层级对象-->
      <TabPane v-if="from !== 'admin'" :label="$t('m_field_resourceLevel')" name="2">
      </TabPane>
      <!--对象组-->
      <TabPane v-if="from !== 'admin'" :label="$t('m_object_group')" name="3">
      </TabPane>
      <!--对象-->
      <TabPane v-if="from !== 'admin'" :label="$t('m_object_design') + '(' + $t('m_ready_only') + ')'" name="4">
      </TabPane>
      <MetricChange slot="extra" ref="metricChangeRef" @reloadData="reloadData"></MetricChange>
    </Tabs>
    <GeneralGroup v-if="activeName === '1'" ref="metricList" @totalCount="setTotalCount"></GeneralGroup>
    <LevelGroup v-if="activeName === '2'" ref="metricList" @totalCount="setTotalCount"></LevelGroup>
    <ObjectGroup v-if="activeName === '3'" ref="metricList" @totalCount="setTotalCount"></ObjectGroup>
    <Object v-if="activeName === '4'" ref="metricList" @totalCount="setTotalCount"></Object>
  </div>
</template>

<script>
import MetricChange from './components/metric-change.vue'
import GeneralGroup from './general-group.vue'
import LevelGroup from './level-group.vue'
import ObjectGroup from './object-group.vue'
import Object from './object.vue'
export default {
  components: {
    GeneralGroup,
    LevelGroup,
    ObjectGroup,
    Object,
    MetricChange
  },
  props: {
    from: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      activeName: this.from === 'admin' ? '1' : '2'
    }
  },
  methods: {
    // 切换区数字
    setTotalCount(val, currentMetric) {
      this.$refs.metricChangeRef.setTotalCount(val, currentMetric)
    },
    // 切换区重新加载数据
    reloadData(metricType) {
      this.$refs.metricList.reloadData(metricType)
    },
  }
}
</script>

<style lang="less" scoped>
.monitor-metric-config {
  width: 100%;
}
</style>
