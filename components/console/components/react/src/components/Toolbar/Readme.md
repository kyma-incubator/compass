```jsx
<div>
  <h5>Toolbar</h5>
  <Toolbar headline="Header" />

  <h5>Toolbar with description</h5>
  <Toolbar headline="Header" description="More details in description" />

  <h5>Toolbar with large back button</h5>
  <Toolbar back={()=>{alert("goBack()")}} largeBackButton headline="Header" />

  <h5>Toolbar with small back button</h5>
  <Toolbar back={()=>{alert("goBack()")}} headline="Header" />

  <h5>Toolbar with search input</h5>
  <Toolbar headline="Header">
    <Search placeholder="Search" searchFunction={() => {
      alert('TO DO: Searching...');
    }} />
  </Toolbar>

  <h5>Toolbar with separator line</h5>
  <Toolbar headline="Header" addSeparator>
    <Search placeholder="Search" searchFunction={() => {
      alert('TO DO: Searching...');
    }} />
  </Toolbar>


  <h5>Toolbar - catalog list</h5>
  <Toolbar
    back={() => {
      alert('goBack()');
    }}
    largeBackButton
    headline="Header"
    description="More details in description"
  >
    <Search placeholder="Search" searchFunction={() => {
      alert('TO DO: Searching...');
    }} />
  </Toolbar>
      
  <h5>Toolbar - catalog details</h5>
  <Toolbar
    back={() => {
      alert('goBack()');
    }}
    headline="Header"
    addSeparator
  >
    <Button primary last>Button</Button>
  </Toolbar>

</div>
```
