<template>
  <div class="hello">
    <h1>{{ msg }}</h1>
    <input v-model="query" @keyup.enter="search" placeholder="Good luck" class="search-box">
    <div class="result">
      <div v-if="hasResult()">
        <div v-for="doc in result" class="doc">
          <p>{{ doc.tweets }}</p>
          <p>{{ doc.date }}</p>
        </div>
      </div>
      <div v-else>
        <p>No result, try again :)</p>
      </div>
    </div>
    <div class="foot">
      Powered by <a href="https://github.com/cosmtrek/violet" target="_blank">violet</a> and Vue.js
    </div>
  </div>
</template>

<script>
export default {
  data() {
    return {
      msg: 'Tweets Search',
      query: '',
      result: [],
    };
  },
  methods: {
    search() {
      const searchApi = `https://t.happyhacking.io/violet/search?query=${this.query}`;
      this.$http.get(searchApi).then((response) => {
        const body = response.body;
        if (body.code === '0') {
          this.result = body.docs;
        } else {
          this.result = 'oops error';
        }
      }, (error) => {
        this.result = error.body;
      });
    },
    hasResult() {
      return this.result && this.result.length > 0;
    },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}

.search-box {
  width: 800px;
  height: 40px;
  padding: 2px 6px;
  font-size: 20px;
}

.result {
  margin: 60px auto;
  max-width: 800px;
}

.doc {
  text-align: left;
  border-bottom: 1px solid #eee;
}

.foot {

}
</style>
