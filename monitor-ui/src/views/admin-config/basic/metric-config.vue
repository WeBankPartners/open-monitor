<!--类型配置-->
<template>
  <div class="admin-metric-config">
    <BasicMetric from="admin" />
  </div>
</template>

<script>
import BasicMetric from '../../metric-config/index'
export default {
  components: {
    BasicMetric
  },
  data() {
    return {}
  },
  methods: {}
}
</script>

<style scoped lang="less">
.admin-metric-config {
  width: 100%;
}
</style>
