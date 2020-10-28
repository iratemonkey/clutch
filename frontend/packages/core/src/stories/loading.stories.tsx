import React from "react";
import type { Meta } from "@storybook/react";

import type { LoadableProps } from "../loading";
import Loadable from "../loading";

export default {
  title: "Core/Loading",
  component: Loadable,
} as Meta;

const Template = (props: LoadableProps) => (
  <Loadable {...props}>
    <div>Testing</div>
  </Loadable>
);

const LargeTemplate = (props: LoadableProps) => (
  <Loadable {...props}>
    <div style={{ justifyContent: "center", width: "1000px" }}>Testing</div>
  </Loadable>
);

export const Primary = Template.bind({});
Primary.args = {
  isLoading: true,
};

export const Overlay = LargeTemplate.bind({});
Overlay.args = {
  ...Primary.args,
  overlay: true,
};
