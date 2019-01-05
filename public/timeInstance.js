// import * as Plotly from "./plotly-finance";

function startTime(){
    var today = moment().format('YYYY-MM-DD HH:mm:ss');

    $.get("/admin/updatetime",function(data,status){
        var last=new Date(data);
        var now=new Date();
        document.getElementById("lastupdate").innerHTML= "Last update time: "+ data;
        if ((now-last)>60000){
            $("#rightnow").css("color","red");
            document.getElementById("rightnow").innerHTML=  "Right now: "+ today;
        }else {
            $("#rightnow").css("color","black");
            document.getElementById("rightnow").innerHTML=  "Right now: "+ today;
        }

    });


    let t = setTimeout(function () {
        startTime()
    }, 500);
}

function ajax_Proc(){
    // $("#what").html("我是一个傻子吧");

    //
    $.get("/admin/data",function(data,status){
        let obj = JSON.parse(data);
        var cols = ['Equipment'];

        for(var item in obj){
            // console.log('item',item);
            for (var col in obj[item]){
                // console.log('col',col);
                if (!cols.includes(col)){
                    cols.push(col)
                }
            }
        }

        var dt = [];
        for(var item01 in obj){
            var tmp = [item01];
            var tmp_arr = cols.slice(1);
            for(var col in tmp_arr){
                tmp.push(obj[item01][tmp_arr[col]]);
            }
            dt.push(tmp);
        }

        Table().init({
            id:'table',
            header:cols,
            data:dt
        });

        //获取table序号，病修改第二列湿度内容
        var x = $("#table");
        x.find("tr").each(function(){
            $(this).find("td:eq(1)").each(function(){
                // console.log($(this).text());
                var h_val = $(this).text();
                if (h_val>90 && h_val<=120){
                    $(this).css("background-color","#FF7575");
                }else if(h_val>80 && h_val<=90){
                    $(this).css("background-color","#F0FFF0");
                }else if (h_val<=80 && h_val>10){
                    $(this).css("background-color","#FFFFCE");
                }else {
                    $(this).css("background-color","#8E8E8E");
                }
            });
        });
    }
);
}



function chart_ProcT(name){
    return function() {
        $.get("/admin/tsdata", function (dt, status) {
                let obj = JSON.parse(dt), TESTER;
                //console.log(obj['E02']['RTS'])
                // var rightnow = moment().format('YYYY-MM-DD HH:mm:ss');
                // console.log(rightnow)

                TESTER = document.getElementById(name);

                let data = [
                    {
                        type: "scatter",
                        mode: "lines",
                        name: 'E02Temp',
                        x: obj['E02']['RTS'],
                        y: obj['E02']['ATS']
                    },
                    {
                        type: "scatter",
                        mode: "lines",
                        name: 'E03Temp',
                        x: obj['E03']['RTS'],
                        y: obj['E03']['ATS']
                    },

                ];
                let layout = {
                    title: '温度时序数据',
                    xaxis: {
                        autorange: true,
                        // range: ['2015-02-17', '2017-02-16'],
                        rangeselector: {
                            buttons: [
                                {
                                    count: 1,
                                    label: '1d',
                                    step: 'day',
                                    stepmode: 'backward'
                                },
                                {
                                    count: 7,
                                    label: '7d',
                                    step: 'day',
                                    stepmode: 'backward'
                                },
                                {step: 'all'}
                            ]
                        },
                        rangeslider: {},
                        type: 'date'
                    },
                    yaxis: {
                        autorange: true,
                        type: 'linear'
                    }
                };
                Plotly.react(TESTER, data, layout);
            }
        );
    }
}


function chart_ProcH(name){
    return function() {

        $.get("/admin/tsdata", function (dt, status) {
                let obj = JSON.parse(dt), TESTER;
                //console.log(obj['E02']['RTS'])
                // var rightnow = moment().format('YYYY-MM-DD HH:mm:ss');
                // console.log(rightnow)

                TESTER = document.getElementById(name);

                let data = [
                    {
                        type: "scatter",
                        mode: "lines",
                        name: 'E02Humi',
                        x: obj['E02']['RTS'],
                        y: obj['E02']['AHS']
                    },
                    {
                        type: "scatter",
                        mode: "lines",
                        name: 'E03Humi',
                        x: obj['E03']['RTS'],
                        y: obj['E03']['AHS']
                    },

                ];
                let layout = {
                    title: '湿度时序数据',
                    xaxis: {
                        autorange: true,
                        rangeselector: {
                            buttons: [
                                {
                                    count: 1,
                                    label: '1d',
                                    step: 'day',
                                    stepmode: 'backward'
                                },
                                {
                                    count: 7,
                                    label: '7d',
                                    step: 'day',
                                    stepmode: 'backward'
                                },
                                {step: 'all'}
                            ]
                        },
                        rangeslider: {},
                        type: 'date'
                    },
                    yaxis: {
                        autorange: true,
                        type: 'linear'
                    }
                };
                Plotly.react(TESTER, data, layout);
            }
        );
    }
}

