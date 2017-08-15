package org.holmesprocessing.totem.services.{$name}

import dispatch.Defaults._
import dispatch.{url, _}
import org.json4s.JsonAST.{JString, JValue}
import org.holmesprocessing.totem.types.{TaskedWork, WorkFailure, WorkResult, WorkSuccess}
import collection.mutable


case class {$name}Work(key: Long, filename: String, TimeoutMillis: Int, WorkType: String, Worker: String, Arguments: List[String]) extends TaskedWork {
  def doWork()(implicit myHttp: dispatch.Http): Future[WorkResult] = {

    val uri = {$name}REST.constructURL(Worker, filename, Arguments)
    val requestResult = myHttp(url(uri) OK as.String)
      .either
      .map({
      case Right(content) =>
        {$name}Success(true, JString(content), Arguments)

      case Left(StatusCode(404)) =>
        {$name}Failure(false, JString("Not found (File already deleted?)"), Arguments)

      case Left(StatusCode(500)) =>
        {$name}Failure(false, JString("Objdump service failed, check local logs"), Arguments) //would be ideal to print response body here

      case Left(StatusCode(code)) =>
        {$name}Failure(false, JString("Some other code: " + code.toString), Arguments)

      case Left(something) =>
        {$name}Failure(false, JString("wildcard failure: " + something.toString), Arguments)
    })
    requestResult
  }
}


case class {$name}Success(status: Boolean, data: JValue, Arguments: List[String], routingKey: String = "{$name}.result.static.totem", WorkType: String = "{$name_toUpper}") extends WorkSuccess
case class {$name}Failure(status: Boolean, data: JValue, Arguments: List[String], routingKey: String = "", WorkType: String = "{$name}") extends WorkFailure


object {$name}REST {
  def constructURL(root: String, filename: String, arguments: List[String]): String = {
    arguments.foldLeft(new mutable.StringBuilder(root+filename))({
      (acc, e) => acc.append(e)}).toString()
  }
}
