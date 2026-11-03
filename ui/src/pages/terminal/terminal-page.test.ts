import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import TerminalToolbar from './components/terminal-toolbar.vue';
import TerminalMobileBar from './components/terminal-mobile-bar.vue';

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

    const trigger = wrapper.find('.theme-dropdown-btn');
    await trigger.trigger('click');

    const items = wrapper.findAll('.theme-item');
    const draculaItem = items.find(item => item.text() === 'Dracula');
    expect(draculaItem).toBeDefined();
    await draculaItem?.trigger('click');

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

  it('emits toggle-mobile-bar event', async () => {
    const wrapper = mount(TerminalToolbar, {
      props: {
        connectionState: 'connected',
        fontSize: 14,
        currentTheme: 'default'
      }
    });

    const toggleBtn = wrapper.find('.toggle-bar-btn');
    await toggleBtn.trigger('click');
    expect(wrapper.emitted('toggle-mobile-bar')).toBeTruthy();
  });
});

describe('TerminalMobileBar Component', () => {
  it('emits key event when ESC and TAB buttons are clicked', async () => {
    const wrapper = mount(TerminalMobileBar, {
      props: {
        isCtrlActive: false,
        isAltActive: false
      }
    });

    const buttons = wrapper.findAll('.key-btn');
    const escBtn = buttons[0];
    const tabBtn = buttons[1];

    await escBtn.trigger('mousedown');
    expect(wrapper.emitted('send-key')?.[0]).toEqual(['\x1b']);

    await tabBtn.trigger('mousedown');
    expect(wrapper.emitted('send-key')?.[1]).toEqual(['\t']);
  });

  it('emits toggle-ctrl and toggle-alt events', async () => {
    const wrapper = mount(TerminalMobileBar, {
      props: {
        isCtrlActive: true,
        isAltActive: false
      }
    });

    const buttons = wrapper.findAll('.toggle-btn');
    const ctrlBtn = buttons[0];
    const altBtn = buttons[1];

    expect(ctrlBtn.classes()).toContain('active');
    expect(altBtn.classes()).not.toContain('active');

    await ctrlBtn.trigger('mousedown');
    expect(wrapper.emitted('toggle-ctrl')).toBeTruthy();

    await altBtn.trigger('mousedown');
    expect(wrapper.emitted('toggle-alt')).toBeTruthy();
  });
});
