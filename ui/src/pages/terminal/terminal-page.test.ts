import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import TerminalToolbar from './components/terminal-toolbar.vue';

describe('TerminalToolbar Component', () => {
  it('renders correct status text based on connectionState prop', async () => {
    const wrapper = mount(TerminalToolbar, {
      props: {
        connectionState: 'connected',
        fontSize: 14,
        currentTheme: 'default'
      }
    });

    expect(wrapper.find('.status-text').text()).toBe('已连接');

    await wrapper.setProps({ connectionState: 'connecting' });
    expect(wrapper.find('.status-text').text()).toBe('连接中...');

    await wrapper.setProps({ connectionState: 'disconnected' });
    expect(wrapper.find('.status-text').text()).toBe('已断开');

    await wrapper.setProps({ connectionState: 'error' });
    expect(wrapper.find('.status-text').text()).toBe('连接错误');
  });

  it('emits font size updates when +/- buttons are clicked', async () => {
    const wrapper = mount(TerminalToolbar, {
      props: {
        connectionState: 'connected',
        fontSize: 14,
        currentTheme: 'default'
      }
    });

    const buttons = wrapper.findAll('.icon-btn');
    const decreaseBtn = buttons[0];
    const increaseBtn = buttons[1];

    await decreaseBtn.trigger('click');
    expect(wrapper.emitted('update:fontSize')?.[0]).toEqual([13]);

    await increaseBtn.trigger('click');
    expect(wrapper.emitted('update:fontSize')?.[1]).toEqual([15]);
  });

  it('emits theme change event', async () => {
    const wrapper = mount(TerminalToolbar, {
      props: {
        connectionState: 'connected',
        fontSize: 14,
        currentTheme: 'default'
      }
    });

    const select = wrapper.find('.theme-select');
    await select.setValue('dracula');

    expect(wrapper.emitted('update:currentTheme')?.[0]).toEqual(['dracula']);
  });

  it('emits reconnect and clear events', async () => {
    const wrapper = mount(TerminalToolbar, {
      props: {
        connectionState: 'disconnected',
        fontSize: 14,
        currentTheme: 'default'
      }
    });

    const reconnectBtn = wrapper.find('.reconnect-btn');
    await reconnectBtn.trigger('click');
    expect(wrapper.emitted('reconnect')).toBeTruthy();

    const clearBtn = wrapper.find('button[title="清空终端屏幕"]');
    await clearBtn.trigger('click');
    expect(wrapper.emitted('clear')).toBeTruthy();
  });
});
