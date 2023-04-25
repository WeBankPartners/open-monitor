<template>
  <div class="m-item">
    <div class="circle" v-if="$attrs.icon">
      <div
        class="inner-circle"
        :style="circleStyle"
      >
        <img class="circle-icon" :src="$attrs.icon" />
      </div>
    </div>
    <div class="item">
      <div class="title">{{ $attrs.title }}</div>
      <div class="value" :class="{['text-' + $attrs.type]: true}">{{ $attrs.value }}</div>
    </div>
  </div>
</template>

<script>
const colorMappings = {
  total: {
    start: '#116ef9',
    end: '#c1d8fa'
  },
  low: {
    start: '#66cc66',
    end: '#9ce89c'
  },
  medium: {
    start: '#f19d38',
    end: '#f1c188'
  },
  high: {
    start: '#da4e2b',
    end: '#f19881'
  }
}
export default {
  name: "circle-item",
  computed: {
    circleStyle() {
      const { type, value, total } = this.$attrs
      const percent = parseInt(value, 10) * 100 / parseInt(total, 10)
      const percent_1 = percent + 0.2

      if (type === 'total') {
        return {
          background: `linear-gradient(135deg, ${colorMappings[type].start} 0%, ${colorMappings[type].end} 100%)`
        }
      }

      return {
        background: `conic-gradient(${colorMappings[type].start} 0, ${colorMappings[type].end} ${percent}%, transparent ${percent_1}%, transparent)`
      }
    }
  }
};
</script>

<style lang="less" scoped>
.m-item {
  display: flex;
  align-items: center;
  margin: auto 30px;

  .circle {
    flex-shrink: 0;
    flex-grow: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    height: 60px;
    width: 60px;
    background: #f2f3f7;
    border-radius: 50%;
    padding: 3px;
    .inner-circle {
      position: relative;
      height: 53px;
      width: 53px;
      flex-shrink: 0;
      flex-grow: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;

      &.total {
        background: linear-gradient(135deg, #116ef9 0%, #c1d8fa 100%);
      }

      &.low {
        background: conic-gradient(
          #66cc66 0,
          #9ce89c 27%,
          transparent 27.2%,
          transparent
        );
      }

      &.medium {
        background: conic-gradient(
          #f19d38 0,
          #f1c188 27%,
          transparent 27.2%,
          transparent
        );
      }

      &.high {
        background: conic-gradient(
          #da4e2b 0,
          #f19881 27%,
          transparent 27.2%,
          transparent
        );
      }

      &::before {
        content: "";
        position: absolute;
        top: 3px;
        left: 3px;
        width: 47px;
        height: 47px;
        border-radius: 50%;
        background: #fff;
      }
    }

    .circle-icon {
      position: absolute;
      width: 24px;
      height: 24px;
    }
  }

  .item {
    margin-left: 10px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    .title {
      font-size: 14px;
      color: #404144;
    }
    .value {
      font-size: 30px;
      &.text-total {
        color: #116ef9;
      }
      &.text-low {
        color: #6fd16e;
      }
      &.text-medium {
        color: #f19d38;
      }
      &.text-high {
        color: #da4e2b;
      }
    }
  }
}
</style>
