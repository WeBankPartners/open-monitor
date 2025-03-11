<template>
  <div class="m-item">
    <vue-ellipse-progress
      v-if="$attrs.icon"
      :progress="progressValue"
      :size="60"
      :empty-thickness="8"
      :legend="false"
      :lineMode="'in -4'"
      :thickness="4"
      :color="gradient"
    >
      <img
        slot="legend-caption"
        class="circle-icon"
        width="24"
        height="24"
        :src="$attrs.icon"
      />
    </vue-ellipse-progress>
    <div class="item">
      <div class="title">{{ $attrs.title }}</div>
      <div class="value" :class="{['text-' + $attrs.type]: true}">
        {{ $attrs.value }}
      </div>
    </div>
  </div>
</template>

<script>
import { VueEllipseProgress } from 'vue-ellipse-progress'

const colorMappings = {
  total: {
    start: '#116ef9',
    end: '#c1d8fa',
  },
  low: {
    start: '#66cc66',
    end: '#9ce89c',
  },
  medium: {
    start: '#f19d38',
    end: '#f1c188',
  },
  high: {
    start: '#da4e2b',
    end: '#f19881',
  },
}
export default {
  name: 'circle-item',
  components: { VueEllipseProgress },
  computed: {
    progressValue() {
      const { value, total } = this.$attrs
      if (isNaN(total) || total === 0) {
        return 0
      }
      return (parseInt(value, 10) * 100) / parseInt(total, 10)
    },
    gradient() {
      const { type } = this.$attrs
      return {
        radial: false,
        colors: [
          {
            color: colorMappings[type].start,
            offset: '0',
          },
          {
            color: colorMappings[type].end,
            offset: '100',
          },
        ],
      }
    }
  },
}
</script>

<style lang="less" scoped>
.m-item {
  display: flex;
  align-items: center;
  margin: auto 40px;

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

@media (max-width: 1680px) {
  .m-item {
    margin: auto 30px;
  }
}

@media (max-width: 1480px) {
  .m-item {
    margin: auto 6px;
  }
}

@media (max-width: 1280px) {
  .m-item {
    margin: auto 10px;
  }
}
</style>
