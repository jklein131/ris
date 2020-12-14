joshk@ joshk $ curl localhost:8080/next -d '{"roll_length":9.0, "include_rush":true}'
{
 "roll_id": 5577006791947779410,
 "length": 6,
 "plan": [
  {
   "component_id": 4,
   "component_size": "3x5",
   "order_date": "2020-10-14T10:10:10Z",
   "position": 1,
   "sku": "RS-MY18-35",
   "rush": true
  },
  {
   "component_id": 7,
   "component_size": "3x5",
   "order_date": "2020-10-16T10:27:30Z",
   "position": 2,
   "sku": "RC-IH17-35",
   "rush": false
  }
 ]
}
joshk@ joshk $ curl localhost:8080/next -d '{"roll_length":9.0, "include_rush":true}'
{
 "roll_id": 8674665223082153551,
 "length": 7,
 "plan": [
  {
   "component_id": 6,
   "component_size": "2.5x7",
   "order_date": "2020-10-16T10:27:30Z",
   "position": 1,
   "sku": "RS-DS55-27",
   "rush": true
  },
  {
   "component_id": 3,
   "component_size": "2.5x7",
   "order_date": "2020-10-14T09:14:30Z",
   "position": 1,
   "sku": "RC-DY25-27",
   "rush": false
  }
 ]
}
joshk@ joshk $ curl localhost:8080/next -d '{"roll_length":9.0, "include_rush":true}'
{
 "roll_id": 6129484611666145821,
 "length": 7,
 "plan": [
  {
   "component_id": 8,
   "component_size": "2.5x7",
   "order_date": "2020-10-17T19:25:00Z",
   "position": 1,
   "sku": "RS-AH27-27",
   "rush": false
  }
 ]
}
joshk@ joshk $ curl localhost:8080/next -d '{"roll_length":9.0, "include_rush":true}'
{
 "roll_id": 4037200794235010051,
 "length": 0,
 "plan": []
}
