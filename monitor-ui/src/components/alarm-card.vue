<template>
  <Card style="margin-bottom: 15px">
    <template #title>
      <div
        class="col-md-10"
        style="padding: 0; color: #404144; font-size: 16px"
      >
        <img class="bg" src="../assets/img/icon_alarm_H_cube.png" />
        {{ data.content }}
      </div>
      <div
        class="col-md-2"
        style="padding: 0; text-align: right; color: #7e8086"
      >
        {{ data.start_string }}
      </div>
    </template>
    <div style="position: absolute; top: 10px; right: 10px">
      <Tooltip :content="$t('menu.endpointView')">
        <Icon
          type="ios-stats"
          size="18"
          class="fa-operate"
          v-if="!data.is_custom"
          @click="goToEndpointView(data)"
        />
      </Tooltip>
      <Tooltip :content="$t('close')">
        <Icon
          type="ios-eye-off"
          size="18"
          class="fa-operate"
          @click="deleteConfirmModal(data, false)"
        />
      </Tooltip>
      <Tooltip :content="$t('m_remark')">
        <Icon
          type="ios-pricetags-outline"
          size="18"
          class="fa-operate"
          @click="remarkModal(data)"
        />
      </Tooltip>
    </div>
    <ul>
      <li>
        <label class="card-label" v-html="$t('field.endpoint')"></label>
        <div class="card-content">
          {{ data.endpoint }}
          <img
            class="filter-icon"
            @click="addParams('endpoint', data.endpoint)"
            src="../assets/img/icon_filter.png"
          />
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('field.metric')"></label>
        <div class="card-content">
          {{ data.s_metric }}
          <img
            class="filter-icon"
            @click="addParams('metric', data.s_metric)"
            src="../assets/img/icon_filter.png"
          />
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('tableKey.tags')"></label>
        <div class="card-content">
          <div class="tags-box">
            <template v-if="data.tags">
              <Tag
                type="border"
                v-for="(t, tIndex) in data.tags.split('^')"
                :key="tIndex"
                >{{ t }}</Tag
              >
            </template>
          </div>
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('details')"></label>
        <div class="card-content">
          <Tag color="default"
            >{{ $t("tableKey.start_value") }}:{{ data.start_value }}</Tag
          >
          <Tag color="default" v-if="data.s_cond"
            >{{ $t("tableKey.threshold") }}:{{ data.s_cond }}</Tag
          >
          <Tag color="default" v-if="data.s_last"
            >{{ $t("tableKey.s_last") }}:{{ data.s_last }}</Tag
          >
          <Tag color="default" v-if="data.path"
            >{{ $t("tableKey.path") }}:{{ data.path }}</Tag
          >
          <Tag color="default" v-if="data.keyword"
            >{{ $t("tableKey.keyword") }}:{{ data.keyword }}</Tag
          >
        </div>
      </li>
    </ul>
  </Card>
</template>

<script>
export default {
  props: {
    data: Object,
  },
  methods: {
    goToEndpointView(alarmItem) {
      const endpointObject = {
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split("_").slice(-1)[0],
      };
      localStorage.setItem("jumpCallData", JSON.stringify(endpointObject));
      this.$router.push({ path: "/endpointView" });
    },
    deleteConfirmModal(rowData, isBatch) {
      this.isBatch = isBatch;
      this.selectedData = rowData;
      this.isShowWarning = true;
    },
    remarkModal(item) {
      this.modelConfig.addRow = {
        id: item.id,
        message: item.custom_message,
        is_custom: false,
      };
      this.$root.JQ("#remark_Modal").modal("show");
    },
    addParams(key, value) {
      this.$parent.filters[key] = value;
      this.$parent.getAlarm();
    },
  },
};
</script>

<style scoped lang="less">
/deep/ .ivu-card-head {
  background: #f2f3f7;
  display: flex;
  align-items: center;
}

/deep/ .ivu-card-body {
  position: relative;
  line-height: 28px;
}

li {
  display: flex;
}
.card-label {
  width: 80px;
  color: #7e8086;
  font-size: 12px;
  text-align: left;
}
.card-content {
  width: ~"calc(100% - 80px)";
  color: #404144;
  overflow: hidden;
  .tags-box {
    width: max-content;
  }
}
.filter-icon {
  position: relative;
  top: -2px;
  margin-left: 6px;
  cursor: pointer;
}

.fa-operate {
  margin: 8px;
  float: right;
  font-size: 16px;
  cursor: pointer;
}
</style>
