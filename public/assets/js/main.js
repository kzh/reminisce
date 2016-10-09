var url = "ws://localhost:8080/ws";
var socket = new WebSocket(url);
socket.onmessage = receive;

var Event = function(line, change) {
    this.line   = line;
    this.change = change;
};

var events = [];
var eventIndex = 0;

var lastMarker;

var editor, mem;
window.onload = function() {
    editor = ace.edit("editor");
    editor.getSession().setMode("ace/mode/javascript");
    editor.setTheme("ace/theme/github");
    editor.getSession().setUseWorker(false);
    editor.setShowPrintMargin(false);

    mem = document.querySelector(".mem > tbody"); 
    reset();

    document.querySelector(".control").onclick = function() {
        if (this.className.indexOf("play") != -1) {
            this.className = "control stop icon";

            var payload = {
                Method: "init",
                Code:   editor.getValue(),
            };

            socket.send(JSON.stringify(payload));
        } else if (this.className.indexOf("stop") != -1) {
            this.className = "control play icon";
            reset();
        }
    };

    document.querySelector(".step").onclick = function() {
        if (eventIndex == events.length) return 

        events[eventIndex].change();
        var line = events[eventIndex].line - 1;

        if (line != -1) { 
            var Range = ace.require('ace/range').Range;

            editor.session.removeMarker(lastMarker)
            lastMarker = editor.session.addMarker(new Range(line, 0, line, 10000), "foo", "line");
        }

        eventIndex++;
    };
};

function receive(msg) {
    var obj = JSON.parse(msg.data);
    console.log(obj);

    var ev = new Event(obj.line, null);
    if (obj.register != "") {
        ev.change = function(reg, change) {
            return function() {
                console.log(reg);
                console.log(change);
                setRegister(reg, change);
            }
        }(obj.register, obj.change);
    } else if (obj.location != -1) {
        ev.change = function(loc, change) {
            return function() {
                console.log(loc);
                console.log(change);
                setMemoryLoc(loc, change);
            }
        }(obj.location, obj.change);
    }

    events.push(ev);
}

function setMemoryLoc(loc, bytes) {
    mem = document.querySelector(".mem > tbody"); 
    loc = 0xc0000000 - loc - 8;
    
    for (var i = 0; i != bytes.length; i++) {
        var row   = (loc - (loc % 16)) / 16;
        var colum = loc % 16;

        mem.children[row].children[colum].innerText = bytes[i];
        loc++;
    }
}

function bytesToHex(bytes) {
    for (var hex = [], i = 0; i < bytes.length; i++) {
        hex.push((bytes[i] >>> 4).toString(16));
        hex.push((bytes[i] & 0xF).toString(16));
    }
    return hex.slice(8, 16).join("");
}

function setRegister(reg, j) {
    var s = bytesToHex(j);

    var els = document.querySelectorAll("tbody.registers tr");
    for (var i = 0; i != els.length; i++) {
        if (els[i].children[0].innerText.indexOf(reg) != -1) {
            els[i].children[1].innerText = "0x" + s;
        }
    }
}

function reset() {
    mem.innerHTML = "";

    var id = 0;
    for (var i = 0; i != 100; i++) {
        var tr = document.createElement("tr"); 
        for (var j = 0; j != 16; j++) {
            var td = document.createElement("td");
            td.id = id++;
            tr.appendChild(document.createElement("td"));
        }
        mem.appendChild(tr);
        editor.session.removeMarker(lastMarker)
    }

    var els = document.querySelectorAll("tbody.registers tr");
    for (var i = 0; i != els.length; i++) {
        els[i].children[1].innerText = "0x00000000";
    }

    events = [];
    eventIndex = 0;    
}
