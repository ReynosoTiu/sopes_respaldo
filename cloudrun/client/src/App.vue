<template>
  <div id="app">
    <h1>Logs de MongoDB</h1>
    <button @click="fetchLogs">Actualizar Logs</button>
    <table>
      <thead>
        <tr>
          <th>Timestamp</th>
          <th>Mensaje</th>
          <th>Error</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="log in filteredLogs" :key="log._id">
          <td>{{ log.timestamp }}</td>
          <td>{{ log.message }}</td>
          <td>{{ log.error }}</td>
        </tr>
      </tbody>
    </table>
    
  </div>
  <div>
    <button @click="prevPage" :disabled="page === 0">Anterior</button>
    <button @click="nextPage" :disabled="page === maxPage">Siguiente</button>
  </div>
</template>
<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'App',
  data() {
    return {
      logs: [],
      searchQuery: '',
      page: 0,
      pageSize: 6
    }
  },
  computed: {
    filteredLogs() {
      return this.logs
        .filter(log => log.message.toLowerCase().includes(this.searchQuery.toLowerCase()))
        .slice(this.page * this.pageSize, (this.page + 1) * this.pageSize);
    },
    maxPage() {
      return Math.ceil(this.logs.length / this.pageSize) - 1;
    }
  },
  methods: {
    async fetchLogs() {
      try {
        const response = await fetch('/logs')
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        this.logs = await response.json()
      } catch (error) {
        console.error('There was an error fetching the logs:', error)
      }
    },
    prevPage() {
      if (this.page > 0) this.page--;
    },
    nextPage() {
      if (this.page < this.maxPage) this.page++;
    }
  },
  mounted() {
    this.fetchLogs()
  }
})
</script>

<style>
#app {
  max-height:80%;
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  margin: 20px auto;
  padding: 20px;
  background-color: #f5f5f5;
  box-shadow: 0 0 10px rgba(0,0,0,0.1);
}

h1 {
  font-size: 26px;
  color: #333;
  margin-bottom: 20px;
}

input {
  padding: 10px;
  margin-right: 10px;
  font-size: 16px;
  border: 2px solid #ccc;
  border-radius: 5px;
  width: 300px;
}

button {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 10px 20px;
  font-size: 16px;
  cursor: pointer;
  border-radius: 5px;
  transition: background-color 0.3s, box-shadow 0.3s;
}

button:hover {
  background-color: #45a049;
  box-shadow: 0 5px 15px rgba(0,0,0,0.2);
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  background-color: white;
  box-shadow: 0 2px 5px rgba(0,0,0,0.1);
}

th,
td {
  border: 1px solid #ddd;
  padding: 8px 12px;
  text-align: left;
}

th {
  background-color: #4CAF50;
  color: white;
}

tr:nth-child(even) {
  background-color: #f9f9f9;
}

tr:hover {
  background-color: #f1f1f1;
}
</style>