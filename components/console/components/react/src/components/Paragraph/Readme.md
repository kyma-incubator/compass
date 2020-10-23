Paragraph is our basic paragraph text element. Use it when you have a lot to say. It has four sizes: `extraSmall`, `small`, `medium` (default), and `large`. It also has four font weights: `light`, `normal` (default), `medium`, and `bold`.

```jsx
<div>
  <h5>EXTRA SMALL</h5>
  <Paragraph modifiers={['extraSmall']}>
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras eu lorem eget erat feugiat sagittis eu at lorem. Fusce et arcu imperdiet dolor interdum lacinia. Mauris eu dictum erat, a convallis ex. Nulla sit amet quam quis nibh facilisis fermentum non vel neque. Vestibulum placerat libero at eros porta scelerisque.
  </Paragraph>

  <h5>SMALL</h5>
  <Paragraph modifiers={['small']}>
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras eu lorem eget erat feugiat sagittis eu at lorem. Fusce et arcu imperdiet dolor interdum lacinia. Mauris eu dictum erat, a convallis ex. Nulla sit amet quam quis nibh facilisis fermentum non vel neque. Vestibulum placerat libero at eros porta scelerisque.
  </Paragraph>

  <h5>DEFAULT</h5>
  <Paragraph>
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras eu lorem eget erat feugiat sagittis eu at lorem. Fusce et arcu imperdiet dolor interdum lacinia. Mauris eu dictum erat, a convallis ex. Nulla sit amet quam quis nibh facilisis fermentum non vel neque. Vestibulum placerat libero at eros porta scelerisque.
  </Paragraph>

  <h5>LARGE</h5>
  <Paragraph modifiers={['large']}>
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras eu lorem eget erat feugiat sagittis eu at lorem. Fusce et arcu imperdiet dolor interdum lacinia. Mauris eu dictum erat, a convallis ex. Nulla sit amet quam quis nibh facilisis fermentum non vel neque. Vestibulum placerat libero at eros porta scelerisque.
  </Paragraph>
</div>
```
