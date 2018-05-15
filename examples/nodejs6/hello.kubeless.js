module.exports = {
    foo: function (event, context) {
        let name = "Noone";
        if (event.data.name) {
          name = event.data.name;
        }
        let place = "Nowhere";
        if (event.data.place) {
          place = event.data.place;
        }
        return {myField: 'Hello, ' + name + ' from ' + place};
    },
};
