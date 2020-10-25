```jsx
<div>
  <h5>Search</h5>
  <Search/>

  <h5>Search with placeholder</h5>
  <Search placeholder="Search"/>

  <h5>Search with `onChange` function (see Developer Console for output)</h5>
  <Search onChange={(e) => console.log("Searching", e.target.value)}/>
</div>
```
