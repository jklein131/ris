# Ruggable Backend Technical

## Background
Ruggable rugs are made to order. This means that generally speaking we don't stock finished goods and instead we stock raw materials that can be turned into a variety of our products. Our main raw material is a roll of fabric that we print designs on. The length of the roll may vary from roll to roll.

There are 3 different sizes of rugs that need to be printed, 2.5x7's (which are considered runners), 3x5's and 5x7's. Below is an example of the orientation of how the rugs would be laid out when printed.
<img src='images/PrintedRugLayoutExample.png' />

This table shows how the different size rugs would be printed.
|Rug Size | Length (ft) | Width (ft) | Side by Side Printing |
| --- | --- | --- | --- |
| 2.5' x 7' | 7 | 2.5 | Yes |
| 3' x 5' | 3 | 5 | No |
| 5' x 7' | 7 | 5 | No |
## Problem
 An operator of a printer needs to know what they should be printing next. We try to maintain a first in, first out (FIFO) approach with the exception of items that need to be rushed. They will utilize a web app that calls your endpoint which will tell them the next highest priority items to print.
## Requirements
### Endpoint
- Your endpoint should return a list of the next items that are to be printed.
- These rugs should be in priority order and the position field should reflect this priority.
    - One exception to this is that runners are not always next to each other priority wise. You can pull a runner from later in the queue to fill an empty space
- Only components with a status of `Pending` should be included
- Only orders that have `cancelled` set to false should be included.
- If the request parameter `include_rush` is selected, rush and non-rushed rugs can be returned in the query.
- If the request parameter `include_rush` is set to false, only non-rushed rugs should be returned in the query.
- The sum of the length of the rugs returned should be less than or equal to the length of the roll.
#### Priority
Components are broken into two "buckets." The first bucket is every component that has `rush=true` and the second bucket is all orders where `rush=false`. Within each bucket, the highest priority items are the oldest orders and the rushed orders have priority over non rushed rugs. Below is an example of the priority would work.
 | Component ID | Rush | Order Date |
 | --- | --- | --- |
 | 99 | True | 2020-12-01 |
 | 125 | True | 2020-12-02 |
 | 133 | True | 2020-12-03 |
 | 27 | False | 2020-10-13 |
 | 30 | False | 2020-11-22 |
 | 55 | False | 2020-11-29 |
 | 128 | False | 2020-12-02|

#### Input
Your endpoint should accept the following inputs:
- `roll_length` (decimal) - The length of the roll being planned for in feet.
- `include_rush` (boolean) - if items that are marked as rush should be included in this plan or not

##### Sample Input
```
{
    "roll_length": 25.62,
    "include_rush": true
}
```

#### Output
The data should be returned as JSON.
Your endpoint should return the following:
- `roll_length` (in feet)
- An array of components on a roll titled `plan`. Each entity in the array should contain:
    - `id` (of the component)
    - `component_size`
    - `order_date`
    - `position` - The position should denote where in the plan a rug would be printed. If a runner (a 2.5x7 rug) is printed side by side, both rugs should have the same position. In the example image above, the 5x7 would be in position 1, both 2.5x7's would be in position 2 and the 3x5 would be in position 3.
    - `sku`
    - `rush`
##### Sample Output
```
{
    "roll_id": 2562,
    "length": 14.2
    "plan":[
        {
            "id": 5683,
            "position": 1,
            "size": "2.5x7",
            "order_date": "2020-10-13 04:27:30-07:00",
            "sku": "RS-1234-27",
            "rush": true
        },
        {
            "id": 2562,
            "position": 1,
            "size": "2.5x7",
            "order_date":"2020-09-14 16:24:24-07:00",
            "sku": "RC-1013-27",
            "rush": false
        },
        {
            "id": 9876,
            "position": 2
            "size": "3x5",
            "order_date":"2020-11-22 10:02:06-07:00",
            "sku": "RS-1234-27",
            "rush": true
        },
        {
            "id": 5684,
            "position": 3
            "size": "3x5",
            "order_date":"2020-11-22 10:30:24-07:00",
            "rush": true
        }
    ]
}
```
## Assumptions
- There is only one width of roll, 5 feet wide.
- A line item has a quantity of 1. If more than one of a particular design is ordered, it will appear as a separate line item

## Supplied Tables
- **`component`**: You can think of a component as a synonym for a rug. There will be one component per line item.
- **`line_item`**: A particular item that was ordered. There are potentially n line items for each order
- **`order`**: Contains information about the order.

There is a `db.sql` file included that will set these tables up with some data that you should be able to use.

## Other Considerations
- Feel free to add tables or columns to existing tables.
- Don't remove any of the existing tables, but if you feel there is a better way to handle a situation, make a note of it.
- Feel free using a modern popular language that you are comfortable with. We use Node at Ruggable but we are more interested in how you approach solving the problem over the specific language.
- Don't worry if you aren't able to fully finish everything in time. Focus mainly on the core logic.
- Please upload your code to Github and share the link with us.


# Algorithm Approach

# Planning algorithm (Not Implimented)
 We're first going to figure out the best print profile given the length suggested. We essentially have two lengths, 7 and 3.
 5*7 and 2.5x7|2.5x7 are both 7ft, and 3x5 is the only odd shape, so a specific priority must be given to 3ft lengths to
 minimize wasted material.

# Design Decisions
- Since this is a planning operation, it does not to be extremely quick

# Missing features from the Algorithm
- Priority Lock
  - If priority rugs are always being submitted, and there isn't enough printers to fullfil the load, non-high priority rugs should "convert" into
    priority rugs if in the queue for over a set period of time (1-2 days). This can be done by adding the comparison in `HighestPriority` function.
    ```golang if time.Now().Add(-time.Hour*24).After(r.OrderDate) || time.Now().Add(-time.Hour*24).After(l.OrderDate) {
	 	return r.OrderDate.After(l.OrderDate)
	 }```
- Optimal Fabric Usage
  - Waste is an important metric when calculating cost, right now the rug planning uses a priority based approach based on order, it should also take into consideration how much material is wasted on a non-planned print. I wrote a simple planner (`UseRugBlocks`) which will optimally determine how many blocks can be printed to best accommodate the space of the rug. For example, if a rug of size 9 is submitted to the `/next` endpoint, it could be fullfiled by 3*3x5 rugs (with zero wasted material) or 1*5x7 (with 2ft of wasted material). Optimal planning can reduce labor and material costs.
- Unfulfilled order timeout
  - As listed below, printer failure is a hard edgecase, there should be an unfulfilled order timeout, where if an order has been scheduled to a roll, but not been confirmed printed or fullfilled, it should be rescheduled within a certain amount of time.
- Order Neighboring
  - To save on packaging and transportation costs, orders with the same destination address should all be printed around the same time in the same factory.
-

# Edge Cases
- Printer Failure
  - If the printer is assigned a roll, and fails to print it, then the roll assigned to that printer needs to be requeued or reprinted.
  - If the printer fails halfway, or some rugs were printed incorrectly, they will need to be requeued or reprinted.
- Order Cancellation
  - What happens if an order has been canceled but product has already been printed, how do we organize that in the database?
- Rug size is wrong
  - A fragment is two small to be printed on, or no product is available to fill the rug portions.
- As mentioned above, Priority Lock is a problem. If priority is overused, some rugs may never be printed.


# Data Considerations About Schema
- No Traceability of datetime makes it hard to make assumptions about how rug production works. When the `component.status` is updated, there is no record of when/why that was updated, this makes it hard to answer questions about our data:
  - What is the average time until an order is cancelled?
  - How long does it take for a rug to be printed after an order is received?
- No Information about users/accounts/permissions, who authorized the request, who has permission to cancel said job and why.



# Observability and Stability
- This application has limited observability into what it's doing, and it's status.

# Testing and Simulation

# Missing edge cases


