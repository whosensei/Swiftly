"use client"
import { useState } from "react";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import axios from "axios";

export function Shorten() {

  const [value, setValue] = useState("");

  async function HandleShorten() {
    const url: string = await axios.get("/");
    setValue(url);
    console.log(value)
  }

  return (
    <div className="flex flex-row gap-2 w-xl">
      <Input
        type="text"
        placeholder="Enter your URL"
        value={value}
        onChange={(e) => setValue(e.target.value)}
      />
      <Button onClick={HandleShorten}>Submit</Button>
    </div>
  );
}
