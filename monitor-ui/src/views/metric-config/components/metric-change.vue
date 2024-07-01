<template>
  <div class="btn-group-custom">
    <div class="btn-group-item" @click="metricTypeChange('originalMetrics')" :class="{'active-metric-type': metricType === 'originalMetrics'}">
      原始指标
    </div>
    <div class="btn-group-item" @click="metricTypeChange('comparisonMetrics')" :class="{'active-metric-type': metricType === 'comparisonMetrics'}">
      同环比指标
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      metricType: 'originalMetrics',
    }
  },
  methods: {
    metricTypeChange (type) {
      this.metricType = type
      this.$emit('reloadData', this.metricType)
    }
  }
}
</script>

<style lang="less" scoped>
.btn-group-item {
  display: inline-block;
  height: 32px;
  line-height: 30px;
  margin: 0;
  padding: 0 15px;
  font-size: 14px;
  color: #515a6e;
  transition: all .2s ease-in-out;
  cursor: pointer;
  border: 1px solid #dcdee2;
  // border-left: 0;
  background: #fff;
}
.active-metric-type {
  background: #2d8cf0;
  color: #fff;
}
.btn-group-custom {
  margin-right: 8px;
} 
.btn-group-custom .btn-group-item:first-child {
  border-radius: 4px 0 0 4px;
}
.btn-group-custom .btn-group-item:last-child {
  border-radius: 0 4px 4px 0;
  position: relative;
  left: -1px;
}
</style>