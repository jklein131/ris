select component.id, component.size, "order".order_date, line_item.sku, line_item.rush from component, line_item, "order"
		where component.line_item_id = line_item.id
		 and line_item.order_id = "order".id
		 and component.status = 'Pending'
		 and  "order".cancelled = 'false'
         and component.size = '3x5'
		ORDER BY line_item.rush DESC, "order".order_date ASC;
	/*
			- `id` (of the component)
		    - `component_size`
		    - `order_date`
		    - `position` - The position should denote where in the plan a rug would be printed. If a runner (a 2.5x7 rug) is printed side by side, both rugs should have the same position. In the example image above, the 5x7 would be in position 1, both 2.5x7's would be in position 2 and the 3x5 would be in position 3.
		    - `sku`
			- `rush`
	*/
