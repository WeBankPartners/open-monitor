<template>
  <div class="metrics-bar" :class="{'single': $attrs.metrics && $attrs.metrics.length === 1}" v-show="$attrs.metrics && $attrs.metrics.length > 0">
    <div
      class="bar-item"
      v-for="(mtc, idx) in $attrs.metrics"
      :key="mtc.name + mtc.type"
      :style="{
        background: barColors[idx % 13],
        height: '15px',
        width: `${(100 * mtc.value) / $attrs.total}%`
      }"
      @click="handleClick(mtc)"
    >
      <Tooltip
        :content="`${mtc.name}: ${mtc.value}`"
        placement="top"
        theme="light"
      >
        <div class="content">&nbsp;&nbsp;</div>
      </Tooltip>
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      barColors: [
        '#DE4B7D',
        '#E57A50',
        '#D8CF6B',
        '#AFC8E4',
        '#002B55',
        '#EC6820',
        '#98B63F',
        '#0199D3',
        '#03519F',
        '#535557',
        '#60C7C4',
        '#A7D9BF',
        '#FFDB3B',
      ],
    }
  },
  methods: {
    handleClick({ name, value }) {
      if (+value > 0) {
        this.$emit('onFilter', {
          key: 'metric',
          value: name
        })
      }
    }
  }
}
</script>

<style scoped lang="less">
.metrics-bar {
  position: absolute;
  // top: 732px;
  top: ~"calc(100% - 10px)";
  width: 750px;
  height: 31px;
  background: #ffffff;
  box-shadow: 0px 8px 15px 0px rgba(17, 110, 249, 0.15);
  border-radius: 15px;
  display: flex;
  padding: 8px 10px;

  &.single {
    .bar-item {
      border-radius: 7px !important;
    }
  }

  .bar-item {
    height: 15px;

    /deep/ .ivu-tooltip {
      width: 100%;
    }

    .content {
      width: 100%;
    }
  }

  .bar-item:nth-child(1) {
    border-radius: 7px 0 0 7px;
  }
  .bar-item:last-child {
    border-radius: 0 7px 7px 0;
  }
}
</style>
