<template>
    <Menu mode="horizontal" :theme="theme1" :active-name="activeName" @on-select="menuChange">
        <div class="logo" @click="routerChange">
            <img src="../../assets/logo.png"/>
            <span>Wecube-monitor</span>
        </div>
        <!-- <MenuItem name="mainView">
            <i class="fa fa-line-chart" aria-hidden="true"></i>
            监控视图
        </MenuItem> -->
        <Submenu name="">
            <template slot="title">
                <i class="fa fa-line-chart" aria-hidden="true"></i>
                监控视图
            </template>
            <MenuItem name="mainView">对象视图</MenuItem>
            <MenuItem name="metricConfig">视图配置</MenuItem>
            <MenuItem name="viewConfigIndex">自定义视图</MenuItem>
        </Submenu>
        <!-- <Submenu name="2">
            <template slot="title">
                <i class="fa fa-television" aria-hidden="true"></i>
                监控告警
            </template>
            <MenuGroup title="使用">
                <MenuItem name="2-1">新增和启动</MenuItem>
                <MenuItem name="2-2">活跃分析</MenuItem>
                <MenuItem name="2-3">时段分析</MenuItem>
            </MenuGroup>
            <MenuGroup title="留存">
                <MenuItem name="2-4">用户留存</MenuItem>
                <MenuItem name="2-5">流失用户</MenuItem>
            </MenuGroup>
        </Submenu> -->
        <MenuItem name="monitorConfigIndex">
            <i class="fa fa-gears" aria-hidden="true"></i>
            监控配置
        </MenuItem>
        <MenuItem name="alarmManagement">
            <i class="fa fa-bell" aria-hidden="true"></i>
            告警管理
        </MenuItem>
    </Menu>
</template>
<script>
    export default {
        data () {
            return {
                theme1: 'dark',
                activeName: ''
            }
        },
        methods: {
            routerChange () {
                if (this.$route.name === 'portal') return
                this.$router.push({ name: 'portal'})
            },
            menuChange (name) {
                this.activeName = name
                if (this.$route.name === name) return
                this.$router.push({ name: name})
            }
        }
    }
</script>
<style lang="less" scoped>
    .logo {
        float: left;
        height: inherit;
        padding-left: 30px;
        cursor: pointer;
        span {
            color: white;
            font-size: 16px;
            font-weight: bolder;
            vertical-align: top;
        }
        img {
            width: 40px; 
            margin: 10px 20px 0;

        }
    }
</style>
