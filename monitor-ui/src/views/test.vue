<template>
  <Table :columns="columns1" :data="data1"></Table>
</template>
<script>
export default {
  data() {
    return {
      columns1: [
        {
          title: 'Name',
          key: 'name'
        },
        {
          title: 'Age',
          key: 'age'
        },
        {
          title: 'Address',
          key: 'address',
          ellipsis: true,
          tooltip: true
        }
      ],
      data1: [
        {
          name: 'John Brown',
          age: 18,
          address: 'New York No. 1 Lake Park',
          date: '2016-10-03'
        },
        {
          name: 'Jim Green',
          age: 24,
          address: 'London No. 1 Lake Park',
          date: '2016-10-01'
        },
        {
          name: 'Joe Black',
          age: 30,
          address: 'Sydney No. 1 Lake Park',
          date: '2016-10-02'
        },
        {
          name: 'Jon Snow',
          age: 26,
          address: 'Ottawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake ParkOttawa No. 2 Lake Park',
          date: '2016-10-04'
        }
      ]
    }
  }
}
</script>
