<!--远程同步-->
<template>
  <div class=" ">
    <div class="w-header" slot="title">
      <div class="title">
        KAFKA
        <span class="underline"></span>
      </div>
    </div>
    <div>
      <Card style="margin: 8px;width: 390px;display: inline-block;" v-for="card in cardList" :key="'card_' + card.id">
        <div slot="title" class="panal-title">
          <Tooltip :content="card.id" max-width="200">
            <h5 class="ellipsis-text">{{ card.id }}</h5>
          </Tooltip>
          <div style="width: 130px;display: inline-block;">
            <div>{{$t('m_update')}}: {{ card.update_user || card.create_user }}</div>
            <div>{{ card.update_time }}</div>
          </div>
        </div>
        <div style="height: 120px">
          <Form ref="formData" :label-width="100">
            <FormItem>
              <span slot="label" style="font-size: 14px;">
                {{ $t('m_http_server') }}：
              </span>
              <Tooltip :content="card.address" transfer max-width="300" style="width: 100%;">
                <div class="text-truncate-">{{ card.address }}</div>
              </Tooltip>
            </FormItem>
          </Form>
          <div class="card-divider"></div>
          <div class="card-content-footer">
            <Button size="small" type="primary" @click.stop="editCard(card)">
              <Icon type="md-create" />
            </Button>
            <Button size="small" type="error" @click.stop="deleteConfirmModal(card)">
              <Icon type="md-trash" />
            </Button>
          </div>
        </div>
      </Card>
      <Card style="margin: 8px;width: 390px;display: inline-block;vertical-align: top;">
        <div slot="title" class="panal-title">
          <h5 class="ellipsis-text">{{$t('m_button_add')}}</h5>
        </div>
        <div style="height: 120px;text-align:center;">
          <Icon @click="addCard" type="md-add-circle" :size=32 style="cursor:pointer;margin-top:42px" :color="'#5384FF'" />
        </div>
      </Card>
    </div>

    <Modal
      :width="600"
      v-model="modelParams.isShow"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{ (modelParams.isAdd ? $t('m_button_add') : $t('m_button_edit')) }}
        </span>
      </div>
      <div>
        <Form ref="formData" :label-width="120">
          <FormItem>
            <span slot="label">
              <span style="color:red">*</span>
              {{ $t('m_configuration_name') }}
            </span>
            <Input v-model="modelParams.params.id" :disabled="!modelParams.isAdd" :maxlength="30" show-word-limit></Input>
          </FormItem>
          <FormItem>
            <span slot="label">
              <span style="color:red">*</span>
              {{ $t('m_http_server') }}
            </span>
            <Input v-model="modelParams.params.address" :placeholder="$t('m_example') + ':http://prometheus-kafka-adapter:8080/receive'"></Input>
          </FormItem>
        </Form>
      </div>
      <div slot="footer">
        <Button :disabled="modelParams.params.id.trim() === '' || modelParams.params.address.trim() === ''" type="primary" @click="saveModal">{{$t('m_button_save')}}</Button>
        <Button @click="cancelModal">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>

    <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      @on-ok="onDeleteConfirm"
      @on-cancel="onCancelDelete"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      cardList: [],
      modelParams: {
        isShow: false,
        isAdd: false,
        params: {
          id: '', // 配置名
          address: '' // http地址
        }
      },
      isShowWarning: false,
      selectedData: {},
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  mounted() {
    this.getList()
  },
  methods: {
    getList() {
      this.request('GET', this.apiCenter.remoteWrite, {}, resp => {
        this.cardList = resp || []
      })
    },
    addCard() {
      this.modelParams.isAdd = true
      this.modelParams.params = {
        id: '', // 配置名
        address: '' // http地址
      }
      this.modelParams.isShow = true
    },
    editCard(item) {
      this.modelParams.isAdd = false
      this.modelParams.params = {
        id: item.id, // 配置名
        address: item.address // http地址
      }
      this.modelParams.isShow = true
    },
    deleteConfirmModal(rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    onDeleteConfirm() {
      this.request('DELETE', this.apiCenter.remoteWrite, this.selectedData, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getList()
      })
    },
    onCancelDelete() {
      this.isShowWarning = false
    },
    saveModal() {
      if (this.modelParams.isAdd) {
        this.request('POST', this.apiCenter.remoteWrite, this.modelParams.params, () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.cancelModal()
          this.getList()
        })
      } else {
        this.request('PUT', this.apiCenter.remoteWrite, this.modelParams.params, () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.cancelModal()
          this.getList()
        })
      }
    },
    cancelModal() {
      this.modelParams.isShow = false
    },
  },
  components: {},
}
</script>
<style scoped lang="less">
.w-header {
  display: flex;
  align-items: center;
  .title {
    font-size: 16px;
    font-weight: bold;
    margin: 0 10px;
    .underline {
      display: block;
      margin-top: -10px;
      margin-left: -6px;
      width: 100%;
      padding: 0 6px;
      height: 12px;
      border-radius: 12px;
      background-color: #c6eafe;
      box-sizing: content-box;
    }
  }
}

.custom-modal-header {
  line-height: 20px;
  font-size: 16px;
  color: #17233d;
  font-weight: 500;
}

.panal-title {
  // display: flex;
  // flex-direction: row;
  // justify-content: space-between;
  color: @blue-2;
  // height: 26px;
}
// .panal-title > div {
//   display: flex;
//   flex-direction: column;
// }

.ellipsis-text {
  width: 210px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  margin-right: 12px;
  // font-weight: 500;
  // line-height: 1.2;
}
.card-divider {
  height: 1px;
  width: 100%;
  background-color: #e8eaec;
}
.card-content-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.text-truncate- {
  word-break: break-all;
  line-height: normal;
  // display: inline-block;
  width: 250px;
  height: 40px;
  display: -webkit-box; /* 使用Webkit的弹性盒子模型显示 */
  -webkit-line-clamp: 2; /* 限制在一个块元素显示的文本的行数 */
  -webkit-box-orient: vertical; /* 设置或检索伸缩盒对象的子元素的排列方式 */
  overflow: hidden; /* 隐藏超出容器的内容 */
  text-overflow: ellipsis; /* 当文本溢出时显示省略号 */
  line-height: 20px; /* 设置行高，确保两行高度适配 */
  max-height: 40px; /* 限制最大高度，配合行高使用，确保显示两行 */
  margin-top: 6px;
}
</style>
