<template>
  <div>
    <Row>
        <Tabs v-model="searchMap.permission" @on-click="onFilterConditionChange()">
            <TabPane :label="$t('m_chart_all')" name="all"></TabPane>
            <TabPane :label="$t('m_can_edit')" name="mgmt"></TabPane>
        </Tabs>
    </Row>
    <div class="chart-list-header mb-3">
        <div class="chart-search">
            <Input
                v-model="searchMap.chartName"
                type="text"
                :placeholder="$t('m_name')"
                clearable
                @on-change="onFilterConditionChange"
            />
            <Input
                v-model="searchMap.chartId"
                type="text"
                :placeholder="$t('m_id')"
                clearable
                @on-change="onFilterConditionChange"
            >
            </Input>
            <Select
                v-model="searchMap.chartType"
                clearable
                filterable
                :placeholder="$t('field.type')"
                @on-change="onFilterConditionChange"
            >
                <Option v-for="item in chartTypeOptionList" :value="item.value" :key="item.value">{{ $t(item.name) }}</Option>
            </Select>
            <Select
                v-model="searchMap.sourceDashboard"
                clearable
                filterable
                :placeholder="$t('m_source_dashboard')"
                @on-change="onFilterConditionChange"
            >
                <Option v-for="item in dashboardOptions" :value="item.id" :key="item.id">{{item.name}}</Option>
            </Select>
            <Select
                v-model="searchMap.useDashboard"
                clearable
                filterable
                multiple
                :placeholder="$t('m_use_dashboard')"
                :max-tag-count="1"
                @on-change="onFilterConditionChange"
            >
                <Option v-for="item in dashboardOptions" :value="item.id" :key="item.id">{{item.name}}</Option>
            </Select>
            <Select
                v-model="searchMap.mgmtRoles"
                clearable
                filterable
                :max-tag-count="1"
                multiple
                :placeholder="$t('m_manage_role')"
                @on-change="onFilterConditionChange"
            >
                <Option v-for="item in mgmtRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
            </Select>
            <Select
                v-model="searchMap.useRoles"
                clearable
                :max-tag-count="1"
                filterable
                multiple
                :placeholder="$t('m_use_role')"
                @on-change="onFilterConditionChange"
            >
                <Option v-for="item in userRolesOptions" :value="item.name" :key="item.name">{{ item.display_name }}</Option>
            </Select>
            <Input
                v-model="searchMap.updateUser"
                type="text"
                :placeholder="$t('m_updatedBy')"
                clearable
                @on-change="onFilterConditionChange"
            >
            </Input>
        </div>
        <Button @click="getChartList" type="primary">{{ $t('m_search') }}</Button>
        <Button @click="resetSearchCondition">{{ $t('m_reset') }}</Button>
    </div>
    <Table
        size="small"
        :columns="chartListColumns"
        :data="tableData"
    />
    <Page
        class="table-pagination mt-3"
        :total="pagination.totalRows"
        @on-change="(currentPage) => {this.pagination.currentPage = currentPage; this.getChartList()}"
        show-sizer
        :current="pagination.currentPage"
        :page-size="pagination.pageSize"
        @on-page-size-change="pageSize => {this.pagination.pageSize = pageSize; this.getChartList()}"
        show-total
    />
    <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveAuth" />
    <Modal
        v-model="isShowWarning"
        :title="$t('delConfirm.title')"
        @on-ok="onDeleteConfirm"
        @on-cancel="onCancelDelete">
        <div class="modal-body" style="padding:30px">
            <div style="text-align:center">
            <p style="color: red">{{$t('delConfirm.tip')}}</p>
            </div>
        </div>
    </Modal>

    <Drawer :title="$t('placeholder.chartConfiguration')" :width="90" :mask-closable="false" v-model="isEditViewShow">
        <editView :chartId="chartId" v-if="isEditViewShow"></editView>
    </Drawer>
    
  </div>
</template>
<script>
import debounce from 'lodash/debounce';
import cloneDeep from 'lodash/cloneDeep'
import isEmpty from 'lodash/isEmpty'
import AuthDialog from '@/components/auth.vue'
import EditView from '@/views/custom-view/edit-view/edit-view'

const initSearchMap = {
    chartName: "",
    chartId: "",
    chartType: "",
    sourceDashboard: "",
    useDashboard: [],
    mgmtRoles: [],
    useRoles: [],
    updateUser: "",
    permission: "all"
}

const initPagination = {
    totalRows: 1,
    currentPage: 1,
    pageSize: 10
}

export default {
  name: '',
  data() {
    return {
        isShowWarning: false,
        searchMap: cloneDeep(initSearchMap),
        request: this.$root.$httpRequestEntrance.httpRequestEntrance,
        pagination: cloneDeep(initPagination),
        userRolesOptions: [],
        mgmtRolesOptions: [],
        dashboardOptions: [],
        chartListColumns: [
            {
                title: this.$t('m_alarmName'),
                align: 'center',
                width: 100,
                key: 'chartName'
            },
            {
                title: this.$t('m_id'),
                align: 'center',
                width: 150,
                key: 'chartId'
            },
            {
                title: this.$t('field.type'),
                key: 'chartType',
                align: 'center',
                width: 90,
                render: (h, params) => {
                    return (
                        <Button
                            type={this.chartTypeMap[params.row.chartType].buttonType}
                            size="small"
                            ghost
                        >{this.$t(this.chartTypeMap[params.row.chartType].label)}</Button>
                    )
                }
            },
            {
                title: this.$t('m_source_dashboard'),
                align: 'center',
                width: 130,
                key: 'sourceDashboard'
            },
            {
                title: this.$t('m_use_dashboard'),
                align: 'center',
                width: 120,
                key: 'useDashboard',
                render: (h, params) => {
                    return (
                        <span>{params.row.useDashboard.join(';')}</span>
                    )
                }
            },
            {
                title: this.$t('m_manage_role'),
                align: 'center',
                width: 100,
                key: 'mgmtRoles',
                render: (h, params) => {
                    return (
                        <span>{params.row.displayMgmtRoles.join(';')}</span>
                    )
                }
            },
            {
                title: this.$t('m_use_role'),
                align: 'center',
                width: 150,
                key: 'useRoles',
                render: (h, params) => {
                    return (
                        <span>{params.row.displayUseRoles.join(';')}</span>
                    )
                }
            },
            {
                title: this.$t('m_updatedBy'),
                align: 'center',
                width: 100,
                key: 'updateUser'
            }, 
            {
                title: this.$t('m_update_time'),
                align: 'center',
                width: 100,
                key: 'updatedTime'
            },
            {
                title: this.$t('m_create_time'),
                align: 'center',
                width: 100,
                key: 'createdTime'
            },
            {
                title: this.$t('table.action'),
                key: 'index',
                align: 'center',
                render: (h, params) => {
                return (params.row.permission === 'mgmt' ? 
                    (<div >
                        <Button size="small" class="mr-1"  type="primary" on-click={() => this.showEditView(params.row)}>
                            <Icon type="md-create" size="16"></Icon>
                        </Button>
                        <Button size="small" type="warning" on-click={() => this.editSingleRoles(params.row)}>
                            <Icon type="md-person" size="16"></Icon>
                        </Button>
                        <Button size="small" type="error" on-click={() => this.showConfirmModal(params.row)}>
                            <Icon type="md-trash" size="16"></Icon>
                        </Button>
                    </div>) : <div>--</div>
                )
                }
            }
        ],
        tableData: [],
        chartTypeOptionList: [
            {
                name: "m_line_chart_s",
                value: "line"
            },
            {
                name: "m_bar_chart",
                value: "bar"
            },
            {
                name: "m_area_chart",
                value: "area"
            }
        ],
        chartTypeMap: {
            line: {
                label: "m_line_chart_s",
                buttonType: "warning"
            },
            bar: {
                label: "m_bar_chart",
                buttonType: "info"
            },
            area: {
                label: "m_area_chart",
                buttonType: "error"
            }
        },
        chartId: "",
        isEditViewShow: false
    }
  },
  mounted(){
    this.pathMap = this.$root.apiCenter.template;
    this.getAllOptions();
    this.getChartList()
  },
  methods: {
    getAllOptions () {
        const params = {
            page: 1,
            size: 1000
        }
        this.request('GET','/monitor/api/v1/user/role/list', params, res => {
            this.userRolesOptions = this.processRolesList(res.data)
        })
        this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
            this.mgmtRolesOptions = this.processRolesList(res);
        }),
        this.request('GET', '/monitor/api/v2/dashboard/all', {}, res => {
            this.dashboardOptions = res;
        })
    },
    processRolesList(list = []) {
        if (isEmpty(list)) return [];
        const resArr = cloneDeep(list).map(item => {
            return {
                ...item,
                key: item.name,
                label: item.display_name
            }
        })
        return resArr
    },
    onFilterConditionChange: debounce(function () {
        this.getChartList()
    }, 300),
    getChartList () {
        const params = Object.assign(cloneDeep(this.searchMap), {
            pageSize: this.pagination.pageSize,
            startIndex: this.pagination.pageSize * (this.pagination.currentPage - 1)
        });
        if (!params.sourceDashboard) {
            params.sourceDashboard = 0;
        }
        if (params.permission === 'all') {
            params.permission = ''
        }
        this.request('POST', '/monitor/api/v2/chart/manage/list', params, responseData => {
            this.tableData = responseData.contents || [];
            this.pagination.totalRows = responseData.pageInfo.totalRows;
        })
    },
    resetSearchCondition() {
        this.searchMap = cloneDeep(initSearchMap);
        this.getChartList()
    },
    showEditView(item) {
        this.chartId = item.chartId;
        this.isEditViewShow = true;
    },
    editSingleRoles(item) {
        this.chartId = item.chartId
        this.$refs.authDialog.startAuth(item.mgmtRoles, item.useRoles, this.mgmtRolesOptions, this.userRolesOptions);
    },
    saveAuth(mgmtRoles, useRoles) {
        const params = {
            chartId: this.chartId,
            mgmtRoles,
            useRoles
        }
        this.request('POST', '/monitor/api/v2/chart/custom/permission', params, () => {
            this.resetPaginationConfig();
            this.getChartList();
        })
    },
    resetPaginationConfig() {
        this.pagination = cloneDeep(initPagination)
    },
    showConfirmModal(item) {
        this.chartId = item.chartId
        this.isShowWarning = true;
    },
    onDeleteConfirm() {
        this.request('DELETE', '/monitor/api/v2/chart/custom', {
            chart_id: this.chartId
        }, () => {
            this.resetPaginationConfig();
            this.getChartList();
        })
    },
    onCancelDelete() {
        this.isShowWarning = false;
    }
  },
  components: {
    AuthDialog,
    EditView
  }
}
</script>

<style scoped lang="less">
.chart-list-header {
    display: flex;
    .chart-search {
        width:90%;
        display: flex;
        flex-wrap: wrap; 
        justify-content: space-between;
        margin-right: 10px;
    }
    .chart-search > div {
        width: 100px;
    }
    .chart-search > div:nth-child(4),
    .chart-search > div:nth-child(5),
    .chart-search > div:nth-child(6),
    .chart-search > div:nth-child(7) {
        width: 180px
    }
}
.table-pagination {
    float: right;
}
</style>
