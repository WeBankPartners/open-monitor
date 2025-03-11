<template>
  <div class="top-stats-container">
    <div class="left">
      <div class="metics-metal">
        <div class="col">
          <div class="title">{{ $t("m_realTimeAlarm") }}</div>
          <img class="time-icon" src="../assets/img/icon_rltm.png" />
        </div>
      </div>
      <div style="display: flex; flex: 1; justify-content: space-around">
        <circle-item
          v-for="stat in $attrs.lstats"
          :key="stat.key"
          :type="stat.type"
          :title="stat.title"
          :total="stat.total"
          :value="$attrs.noData ? '-' : stat.value"
          :icon="stat.icon"
        />
      </div>
    </div>
    <div class="right">
      <div class="metics-metal">
        <div class="col">
          <div class="title">{{ $attrs.rtitle }}</div>
          <img class="time-icon" src="../assets/img/icon_today.png" />
        </div>
      </div>
      <div style="display: flex; justify-content: space-around; flex: 1">
        <circle-item
          v-for="stat in $attrs.rstats"
          :key="stat.key"
          :type="stat.type"
          :title="stat.title"
          :total="stat.total"
          :value="stat.value"
        />
      </div>
    </div>
  </div>
</template>

<script>
import CircleItem from '../components/circle-item.vue'
export default {
  components: {
    CircleItem,
  }
}
</script>

<style scoped lang="less">
.top-stats-container {
  width: 100%;
  height: 90px;
  background: #ffffff;
  border: 2px solid #f2f3f7;
  border-radius: 4px;
  display: flex;

  .metics-metal {
    height: 100%;
    background: linear-gradient(
      90deg,
      #f5f8fe 0%,
      rgba(234, 242, 253, 0) 100%
    );

    .col {
      position: relative;
      width: 120px;
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: flex-start;
      justify-content: center;

      .title {
        font-size: 16px;
      }

      .time-icon {
        width: 32px;
        height: 32px;
        margin-top: 14px;
      }

      &::after {
        content: "";
        position: absolute;
        width: 2px;
        height: 63px;
        right: 0;
        background: #f2f3f7;
      }
    }
  }

  .left {
    width: 42vw;
    // flex-basis: 40%;
    height: 100%;
    display: flex;
  }
  .right {
    width: 58vw;
    // flex-basis: 60%;
    height: 100%;
    display: flex;
    justify-content: flex-start;
    align-items: center;
  }
}
</style>
