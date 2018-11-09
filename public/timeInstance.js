
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


    t=setTimeout(function(){startTime()},500);
}

function ajax_Proc(){
    // $("#what").html("我是一个傻子吧");
    $.get("/admin/data",function(data,status){
        let obj = JSON.parse(data);
        var cols = ['Equipment'];

        for(var item in obj){
            // console.log('item',item);
            for (var col in obj[item]){
                console.log('col',col);
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




