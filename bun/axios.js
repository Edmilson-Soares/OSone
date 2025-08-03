import axios from 'axios'



try {
    const {data}= await axios.post("http://localhost:1337/api/services",{
    "service": "test.create",
    data:{
        "name": "test",
        "description": "test",
        "price": "10.99",
    }
}, {
    headers: {
             apikey: "8rQJ0Afl1sY3NfC57b67X4sgrciuUoVWFtKO4IZQSH_4HsfkVb8yYJAXfCF_nyJTdIpG0C1_IcCfG_Jn_5fJzj2W5osg3A90nPnq3cMi2Oysj8SNU_yuNSDfVw=="
    }
})

console.log(data)
} catch (error) {
    console.log(error)
}


