<template>
    <div class="tree-menu-comm tree-menu">
      <recursiveContainerItem :data="data"></recursiveContainerItem>
    </div>
</template>

<script>
import recursiveContainerItem from '@/views/recursive-view/recursive-container-item'
export default {
  name: '',
  data() {
    return {
      data: [
        {
            "display_name":"name6",
            "charts":null,
            "children":[
                {
                    "display_name":"name5",
                    "charts":null,
                    "children":[
                        {
                            "display_name":"name4",
                            "charts":null,
                            "children":[
                                {
                                    "display_name":"name3",
                                    "charts":null,
                                    "children":[
                                        {
                                            "display_name":"name2",
                                            "charts":null,
                                            "children":[
                                                {
                                                    "display_name":"name1",
                                                    "charts":[
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "mem.used.percent"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"avg"
                                                        },
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "cpu.used.percent"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"p95"
                                                        },
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "load.1min"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"avg"
                                                        }
                                                    ],
                                                    "children":null
                                                },
                                                {
                                                    "display_name":"name8",
                                                    "charts":[
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "mem.used.percent"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"avg"
                                                        },
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "cpu.used.percent"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"p95"
                                                        },
                                                        {
                                                            "id":0,
                                                            "col":0,
                                                            "endpoint":[
                                                                "VM_0_5_centos_192.168.0.5_host",
                                                                "VM_0_16_centos_192.168.0.16_host"
                                                            ],
                                                            "metric":[
                                                                "load.1min"
                                                            ],
                                                            "url":"",
                                                            "aggregate":"avg"
                                                        }
                                                    ],
                                                    "children":null
                                                }
                                            ]
                                        }
                                    ]
                                }
                            ]
                        }
                    ]
                }
            ]
        }
      ]
    }
  },
  // props:{
  //   data:{
  //     type:Array
  //   }
  // },
  methods: {},
  components: {
    recursiveContainerItem
  },
}
</script>

<style scoped lang="less">
  ul {
    padding: 0;
    margin: 0;
    list-style: none;
  }
 
  .tree-menu {
    height: 100%;
    padding: 0px 12px;
    border-right: 1px solid #e6e9f0;
  }

  .tree-menu-comm span {
    display: block;
    font-size: 12px;
    position: relative;
  }

  .tree-menu-comm span strong {
    display: block;
    width: 82%;
    position: relative;
    line-height: 22px;
    padding: 2px 0;
    padding-left: 5px;
    color: #161719;
    font-weight: normal;
  }

  .tree-title {
    border: 1px solid #fff;
    margin-top: 1px;
  }
  .tree-border {
    border: 1px solid #9966;
    padding: 8px;
    margin: 8px;;
  }
</style>

