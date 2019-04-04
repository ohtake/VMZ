const secsPerFragment = 10;
let videoId = 0;

let jumpRequested = false;
let jumpSec = 0;

let csvId = 0;
let actions = [];

const suspiciousActions = new Set([
    "using computer",
    "stretching arm",
]);

function getVideoElement() {
    const id = (videoId % 2 === 0) ? "videoEven" : "videoOdd";
    return document.getElementById(id);
}

function getPreloadVideoElement() {
    const id = (videoId % 2 === 1) ? "videoEven" : "videoOdd";
    return document.getElementById(id);
}

function getCurrentSec() {
    const eVideo = getVideoElement();
    const currentTime = eVideo.currentTime;
    return secsPerFragment * videoId + Math.floor(currentTime);
}

function playNext() {
    videoId++;
    const eVideo = getVideoElement();
    eVideo.play();
    eVideo.style.display = "";
    const eVideoPreload = getPreloadVideoElement();
    eVideoPreload.setAttribute("src", "./video?id=" + (videoId + 1));
    eVideoPreload.style.display = "none";
}

function jumpToHandler(ev) {
    if (!jumpRequested) return;
     if (this === getVideoElement()) {
        jumpRequested = false;
        this.currentTime = jumpSec;
        this.play();
    }
}

function jumpTo(sec) {
    const id = Math.floor(sec / secsPerFragment);
    const secInFragment = sec % secsPerFragment;

    jumpSec = secInFragment;
    jumpRequested = true;
    if (videoId !== id) {
        videoId = id;
        const eVideoPreload = getPreloadVideoElement();
        eVideoPreload.pause();
        eVideoPreload.style.display = "none";
        const eVideo = getVideoElement();
        eVideo.setAttribute("src", "./video?id=" + videoId);
        eVideo.style.display = "";
        eVideoPreload.setAttribute("src", "./video?id=" + (videoId + 1));
    } else {
        // Did not modify video elements. We have to call the envet handler manually.
        jumpToHandler.call(getVideoElement());
    }
}

function updateActions() {
    const currentSec = getCurrentSec();
    let scale = d3.scaleLinear().domain([0, 1]).range([0, 500]);
    const arr = Array.from(actions[currentSec], e => [e[0], e[1], suspiciousActions.has(e[0])]);
    arr.sort((a, b) => b[1] - a[1]);
    const data = arr.slice(0, 5);
    d3.select("#actions").selectAll("div").remove().exit().data(data).enter().append("div")
        .style("width", d => scale(d[1]) + "px")
        .attr("class", d => d[2] ? "suspicious" : "")
        .append("span").text(d => d[0]+ " (" + d[1]*100 + "%)");
}

const tlScaleX = d3.scaleBand().range([0, 800]).padding(0);
const tlScaleY = d3.scaleLinear().domain([1, 0]).range([0, 300]);

function initTimeline() {
    const svg = d3.select("#timeline").append("svg").attr("width", 830).attr("height", 330);
    svg.append("g").attr("class", "bars").attr("transform", "translate(30, 10)");
    svg.append("g").attr("class", "marker").attr("transform", "translate(30, 10)");
    svg.append("g").attr("class", "axisX").attr("transform", "translate(30, 310)");
    svg.append("g").attr("transform", "translate(30, 10)").call(d3.axisLeft(tlScaleY));
    svg.on("click", function() {
        let pos = d3.mouse(this);
        let posX = pos[0];
        let posY = pos[1];
        // Cannot invert?
        let sec = (posX - 30) / tlScaleX.step();
        if (sec >= 0) jumpTo(sec);
    })
}

function updateTimeline() {
    const ss = Array.from(suspiciousActions);
    function suspiciousEvaluator(m) {
        let score = 0;
        ss.forEach(a => {
            score += m.get(a);
        });
        return score;
    }
    let data = actions.map((m, i) => [i, suspiciousEvaluator(m)]);
    tlScaleX.domain(data.map(d => d[0]));
    const svg = d3.select("#timeline svg");
    svg.select("g.bars").selectAll("rect").remove().exit().data(data).enter().append("rect")
        .attr("class", "bar")
        .attr("x", d => tlScaleX(d[0]))
        .attr("y", d => tlScaleY(d[1]))
        .attr("width", tlScaleX.bandwidth())
        .attr("height", d => 300 - tlScaleY(d[1]))
        ;
    updateTimelineCurrent();
    svg.select("g.axisX").call(d3.axisBottom(tlScaleX));
}

function updateTimelineCurrent() {
    const currentSec = getCurrentSec();
    const svg = d3.select("#timeline svg");
    svg.select("g.marker").selectAll("rect").remove().exit().data([currentSec]).enter().append("rect")
        .attr("class", "current")
        .attr("x", d => tlScaleX(d) + tlScaleX.step() / 2)
        .attr("y", tlScaleY(1))
        .attr("width", "1px")
        .attr("height", 300 - tlScaleY(1))
        ;
}

function parseActionCsv(content) {
    const lines = content.split("\n");
    const mapping = new Array(10); // column index -> time
    lines[0].split(',').forEach((v, i) => {
        if (i===0) return; // msut be "label"
        mapping[i - 1] = parseInt(v.substring(1)) - 1;
    })
    const result = new Array(10); // time -> obj
    for (let i = 0; i < result.length; i++) {
        result[i] = new Map(); // label -> softmax value
    }
    for (let i = 1; i < lines.length; i++) {
        const cols = lines[i].split(",");
        const label = cols[0];
        for (let j = 1; j < cols.length; j++) {
            result[mapping[j-1]].set(label, parseFloat(cols[j]));
        }
    }
    return result;
}

function fetchCsv() {
    let text = "";
    const xhr = new XMLHttpRequest();
    xhr.addEventListener("load", function(){
        if (xhr.status === 200) {
            const actions2 = parseActionCsv(xhr.responseText);
            actions = actions.concat(actions2);
            csvId++;
            updateTimeline();
            setTimeout(fetchCsv, 1000);
        } else if (xhr.status === 404) {
            setTimeout(fetchCsv, 5000);
        } else {
            console.error(xhr);
        }
    });
    xhr.open("GET", "/actions?id=" + csvId);
    xhr.send();
}

window.addEventListener("load", function(ev) {
    [getVideoElement(), getPreloadVideoElement()].forEach((v) => {
        v.addEventListener("timeupdate", (evUpdate)=> {
            updateTimelineCurrent();
            updateActions();
        });
        v.addEventListener("ended", playNext);
        v.addEventListener("canplay", jumpToHandler);
    });

    fetchCsv();
    initTimeline();
});
