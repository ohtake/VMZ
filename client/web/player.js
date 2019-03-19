const secsPerFragment = 10;
let videoId = 0;

let jumpRequested = false;
let jumpSec = 0;

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
    const data1 = [
        ["action 1", 0.6, false],
        ["action 2", 0.2, false],
        ["action 3", 0.05, false],
        ["action 4", 0.03, false],
        ["action 5", 0.01, false],
    ];
    const data2 = [
        ["suspicious action 1", 0.5, true],
        ["action 2", 0.2, false],
        ["suspicious action 3", 0.1, true],
        ["action 4", 0.05, false],
        ["action 5", 0.03, false],
    ];
    const data = (currentSec % 2 === 0) ? data1 : data2;
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
    let data = [
        [0, 0.01],
        [1, 0.02],
        [2, 0.03],
        [3, 0.50],
        [4, 0.03],
        [5, 0.01],
        [6, 0.02],
        [7, 0.80],
        [8, 0.03],
        [9, 0.01],
        [10, 0.01],
        [11, 0.02],
        [12, 0.03],
        [13, 0.50],
        [14, 0.03],
        [15, 0.01],
        [16, 0.02],
        [17, 0.80],
        [18, 0.03],
        [19, 0.01],
    ];
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

window.addEventListener("load", function(ev) {
    [getVideoElement(), getPreloadVideoElement()].forEach((v) => {
        v.addEventListener("timeupdate", (evUpdate)=> {
            updateTimelineCurrent();
            updateActions();
        });
        v.addEventListener("ended", playNext);
        v.addEventListener("canplay", jumpToHandler);
    });

    // TODO
    updateActions();
    initTimeline();
    updateTimeline();
});
