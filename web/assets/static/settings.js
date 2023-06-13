// Time-stamp: <2023-06-13 11:32:02 krylon>
// -*- mode: javascript; coding: utf-8; -*-
// Copyright 2020 Benjamin Walkenhorst <krylon@gmx.net>

"use strict";

var settings = {
    "beacon": {
        "active": false,
        "interval": 1000,
    },

    "messages": {
        "queryEnabled": true,
        "interval": 5000,
        "maxShow": 25,
    },

    "items": {
        "hideboring": false,
        "page": 50,
    },

    "chart": {
        "period": 86400, // most recent data to render, age in seconds
    },
};

function initSettings() {
    let item = null;
    
    settings.beacon.active =
        JSON.parse(localStorage.getItem("beacon.active")) ? true : false;
    
    item = JSON.parse(localStorage.getItem("beacon.interval"));
    if (Number.isInteger(item)) {
        settings.beacon.interval = item;
    }

    item = JSON.parse(localStorage.getItem("chart.period"))
    if (Number.isInteger(item)) {
        settings.chart.period = item
    }
} // function initSettings()

function saveSetting(category, attribute, newValue) {
    var cat = settings[category];
    if (cat == undefined) {
        console.log("Invalid category: " + category);
        return;
    }

    var att = cat[attribute];
    if (att == undefined) {
        console.log("Invalid attribute: " + attribute);
        return;
    }

    var key = category + "." + attribute;
    localStorage.setItem(key, newValue);
    settings[category][attribute] = newValue;
} // function saveSetting(group, member, newValue)

// const save_period = () => {
//     const hstr = $("#period")[0].value
//     const hours = Number.parseInt(hstr)

//     saveSetting("chart", "period", hours * 3600)
// }
