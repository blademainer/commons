package xml

import (
	"encoding/xml"
	"fmt"
	"testing"
	"time"
)

func Example() {
	type User struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA
		FromUserName CDATA
		CreateTime   int64
		MsgType      CDATA
		Content      CDATA
	}
	msg := User{
		ToUserName:   "userId",
		FromUserName: "appId",
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      "some message like <hello>"}

	b, _ := xml.MarshalIndent(msg, "", "    ")
	fmt.Println(string(b))
}

func TestCDATA(t *testing.T) {
	type User struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   CDATA    `xml:"to_user_name"`
		FromUserName CDATA    `xml:"from_user_name"`
		CreateTime   int64    `xml:"create_time"`
		MsgType      CDATA    `xml:"msg_type"`
		Content      CDATA    `xml:"content"`
	}
	msg := User{
		ToUserName:   "userId",
		FromUserName: "appId",
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      "some message like <hello>"}

	b, _ := xml.MarshalIndent(msg, "", "    ")
	//<xml>
	//    <to_user_name><![CDATA[userId]]></to_user_name>
	//    <from_user_name><![CDATA[appId]]></from_user_name>
	//    <create_time>1547449470</create_time>
	//    <msg_type><![CDATA[text]]></msg_type>
	//    <content><![CDATA[some message like <hello>]]></content>
	//</xml>
	fmt.Println(string(b))
	user := &User{}
	if e := xml.Unmarshal(b, user); e != nil {
		panic(e.Error())
	}
	fmt.Println(user)
}
