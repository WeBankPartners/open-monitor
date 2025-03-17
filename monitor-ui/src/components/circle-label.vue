<template>
  <div class="cir" :class="{[$attrs.data.type]: true}" :style="cirStyle" @click="onCircleTextClick">
    <div class="text" :style="textStyle">
      <div class="title">{{ $attrs.data.label }}</div>
      <div class="value">{{ percentage }}%</div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'circle-label',
  data() {
    return {
      L: 240
    }
  },
  computed: {
    percentage() {
      const { value, total } = this.$attrs.data
      return ((parseInt(value, 10) * 100 / parseInt(total, 10)) || 0).toFixed(2)
    },
    cirStyle() {
      const { deg, tx, ty } = this.$attrs.data

      return {
        transform: `rotate(${deg}) translate(${2 * tx * this.circleWidth}px, ${ty * this.circleWidth}px)`
      }
    },
    textStyle() {
      const { deg } = this.$attrs.data

      if (deg[0] === '-') {
        return {
          transform: `rotate(${deg.slice(1)})`
        }
      }

      return {
        transform: `rotate(-${deg})`
      }
    },
    circleWidth() {
      const { value, total } = this.$attrs.data
      return Math.max(this.L * ((parseInt(value, 10) / parseInt(total, 10)) || 0), 30)
    },
  },
  methods: {
    onCircleTextClick() {
      if (this.$attrs.data.key) {
        this.$emit('onFilter', {
          key: 'priority',
          value: this.$attrs.data.key
        })
      }
    }
  }
}
</script>

<style scoped lang="less">
.cir {
  position: absolute;
  width: 16px;
  height: 16px;
  background: #FFFFFF;
  border-radius: 50%;
  z-index: 101;
  cursor: pointer;
  .text {
    .title {
      font-size: 16px;
      color: #404144;
    }
    .value {
      font-size: 30px;
      font-weight: 500;
    }
  }

  &.low {
    border: 2px solid #6ED06D;

    .text {
      margin-left: 55px;
      // margin-top: -90px;
      margin-top: -130px;

      .value {
        color: #6ED06D;
      }
    }
  }
  &.mid {
    border: 2px solid #F19D38;

    .text {
      margin-left: 10px;
      margin-top: -90px;

      .value {
        color: #F19D38;
      }
    }
  }
  &.high {
    border: 2px solid #DA4E2B;

    .text {
      margin-left: -50px;
      margin-top: 58px;
      // margin-left: 10px;
      // margin-top: 10px;

      .value {
        color: #DA4E2B;
      }
    }
  }
}
</style>
